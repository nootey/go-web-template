package middleware

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"go-web-template/internal/config"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// Context key for user ID
type ctxKey int

const userIDKey ctxKey = 0

var (
	ErrTokenExpired = errors.New("token has expired")
)

type WebClientUserClaim struct {
	UserID string `json:"ID"`
	jwt.RegisteredClaims
}

type AuthMiddlewareInterface interface {
	WebClientAuthentication(next http.Handler) http.Handler
	GenerateLoginTokens(userID int64, rememberMe bool) (string, string, error)
	SetLoginCookies(w http.ResponseWriter, accessToken, refreshToken string, rememberMe bool)
	ClearLoginCookies(w http.ResponseWriter)
}

type AuthMiddleware struct {
	cfg    *config.Config
	logger *zap.Logger
}

func NewAuthMiddleware(cfg *config.Config, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		cfg:    cfg,
		logger: logger,
	}
}

var _ AuthMiddlewareInterface = (*AuthMiddleware)(nil)

func (m *AuthMiddleware) WebClientAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try access token first
		if accessCookie, err := r.Cookie("access"); err == nil && accessCookie.Value != "" {
			claims, err := m.decodeWebClientToken(accessCookie.Value, "access")
			if err == nil {
				userID, err := m.decodeWebClientUserID(claims.UserID)
				if err == nil {
					// Store user ID in context and continue
					ctx := context.WithValue(r.Context(), userIDKey, userID)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// If access missing/expired, try refresh token
		refreshCookie, err := r.Cookie("refresh")
		if err != nil || refreshCookie.Value == "" {
			m.respondUnauthorized(w, "unauthenticated")
			return
		}

		rClaims, err := m.decodeWebClientToken(refreshCookie.Value, "refresh")
		if err != nil {
			m.respondUnauthorized(w, "unauthenticated")
			return
		}

		userID, err := m.decodeWebClientUserID(rClaims.UserID)
		if err != nil {
			m.respondUnauthorized(w, "unauthenticated")
			return
		}

		// Issue new access token
		if err := m.issueAccessCookie(w, userID); err != nil {
			m.respondUnauthorized(w, "unauthenticated")
			return
		}

		// Store user ID in context and continue
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) GenerateLoginTokens(userID int64, rememberMe bool) (string, string, error) {
	var expiresAt time.Time
	if rememberMe {
		expiresAt = time.Now().Add(m.cfg.Auth.RefreshTTLLong)
	} else {
		expiresAt = time.Now().Add(m.cfg.Auth.RefreshTTLShort)
	}

	accessToken, err := m.generateToken("access", time.Now().Add(m.cfg.Auth.AccessTTL), userID)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := m.generateToken("refresh", expiresAt, userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}

func (m *AuthMiddleware) SetLoginCookies(w http.ResponseWriter, accessToken, refreshToken string, rememberMe bool) {
	var refreshMaxAge int
	if rememberMe {
		refreshMaxAge = int(m.cfg.Auth.RefreshTTLLong.Seconds())
	} else {
		refreshMaxAge = int(m.cfg.Auth.RefreshTTLShort.Seconds())
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    accessToken,
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   int(m.cfg.Auth.AccessTTL.Seconds()),
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    refreshToken,
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   refreshMaxAge,
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *AuthMiddleware) ClearLoginCookies(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    "",
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   -1,
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})

	http.SetCookie(w, &http.Cookie{
		Name:     "refresh",
		Value:    "",
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   -1,
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *AuthMiddleware) issueAccessCookie(w http.ResponseWriter, userID int64) error {
	accessExp := time.Now().Add(m.cfg.Auth.AccessTTL)
	token, err := m.generateToken("access", accessExp, userID)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "access",
		Value:    token,
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   int(time.Until(accessExp).Seconds()),
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

func (m *AuthMiddleware) EncodeWebClientUserID(userID int64) (string, error) {
	key := m.cfg.Auth.EncodeIDSecret
	if len(key) != 32 {
		return "", fmt.Errorf("encryption key must be 32 bytes long for AES-256")
	}

	userIDString := strconv.FormatInt(userID, 10)

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(userIDString), nil)
	ciphertext = append(nonce, ciphertext...) // Prepend nonce to ciphertext

	encoded := base64.StdEncoding.EncodeToString(ciphertext)
	return encoded, nil
}

func (m *AuthMiddleware) decodeWebClientUserID(encodedString string) (int64, error) {
	key := m.cfg.Auth.EncodeIDSecret
	if len(key) != 32 {
		return 0, fmt.Errorf("encryption key must be 32 bytes long for AES-256")
	}

	decodedBytes, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		return 0, err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return 0, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return 0, err
	}

	nonceSize := gcm.NonceSize()
	if len(decodedBytes) < nonceSize {
		return 0, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := decodedBytes[:nonceSize], decodedBytes[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return 0, err
	}

	intUserID, err := strconv.ParseInt(string(plaintext), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse user ID: %v", err)
	}

	return intUserID, nil
}

func (m *AuthMiddleware) generateToken(tokenType string, expiration time.Time, userID int64) (string, error) {
	var jwtKey []byte
	issuedAt := time.Now()

	switch tokenType {
	case "access":
		jwtKey = []byte(m.cfg.Auth.AccessSecret)
	case "refresh":
		jwtKey = []byte(m.cfg.Auth.RefreshSecret)
	default:
		return "", fmt.Errorf("unsupported token type: %s", tokenType)
	}

	encryptedUserID, err := m.EncodeWebClientUserID(userID)
	if err != nil {
		return "", err
	}

	claims := WebClientUserClaim{
		UserID: encryptedUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiration),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			Issuer:    "go-web-template",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (m *AuthMiddleware) decodeWebClientToken(tokenString string, cookieType string) (*WebClientUserClaim, error) {
	var secret string

	switch cookieType {
	case "access":
		secret = m.cfg.Auth.AccessSecret
	case "refresh":
		secret = m.cfg.Auth.RefreshSecret
	default:
		return nil, fmt.Errorf("unknown cookieType: %s", cookieType)
	}

	secretKey := []byte(secret)

	claims := &WebClientUserClaim{}

	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return secretKey, nil
	})

	switch {
	case token.Valid:
		return claims, nil
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		return nil, ErrTokenExpired
	default:
		if err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (m *AuthMiddleware) respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)

	err := json.NewEncoder(w).Encode(map[string]string{
		"title":   "Unauthorized",
		"message": message,
	})
	if err != nil {
		m.logger.Error("failed to encode unauthorized response", zap.Error(err))
	}
}

func GetUserID(r *http.Request) (int64, bool) {
	userID, ok := r.Context().Value(userIDKey).(int64)
	return userID, ok
}
