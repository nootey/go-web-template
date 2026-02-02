package middleware_test

import (
	"encoding/json"
	"fmt"
	"go-web-template/internal/config"
	"go-web-template/internal/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	middleware *middleware.AuthMiddleware
	cfg        *config.Config
}

func (suite *AuthMiddlewareTestSuite) SetupTest() {
	_ = godotenv.Load("../../.env")
	if err := config.Load(); err != nil {
		panic(err)
	}
	suite.cfg = config.Get()

	// Create middleware with short TTLs for testing
	logger, _ := zap.NewDevelopment()
	suite.middleware = middleware.NewAuthMiddleware(
		suite.cfg,
		logger,
		2*time.Second,
		5*time.Second,
		1*time.Minute,
	)
}

func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

func (suite *AuthMiddlewareTestSuite) decodeToken(token, tokenType string) (*middleware.WebClientUserClaim, error) {
	var secret string
	switch tokenType {
	case "access":
		secret = suite.cfg.Auth.AccessSecret
	case "refresh":
		secret = suite.cfg.Auth.RefreshSecret
	default:
		return nil, fmt.Errorf("unknown token type")
	}

	claims := &middleware.WebClientUserClaim{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	return claims, err
}

func (suite *AuthMiddlewareTestSuite) TestGenerateLoginTokens_Success() {
	userID := int64(123)

	accessToken, refreshToken, err := suite.middleware.GenerateLoginTokens(userID, false)

	suite.NoError(err)
	suite.NotEmpty(accessToken)
	suite.NotEmpty(refreshToken)
	suite.NotEqual(accessToken, refreshToken)

	// Verify access token expiration (2 seconds)
	accessClaims, err := suite.decodeToken(accessToken, "access")
	suite.NoError(err)
	suite.WithinDuration(time.Now().Add(2*time.Second), accessClaims.ExpiresAt.Time, 2*time.Second)

	// Verify refresh token expiration (5 seconds for rememberMe=false)
	refreshClaims, err := suite.decodeToken(refreshToken, "refresh")
	suite.NoError(err)
	suite.WithinDuration(time.Now().Add(5*time.Second), refreshClaims.ExpiresAt.Time, 2*time.Second)
}

func (suite *AuthMiddlewareTestSuite) TestGenerateLoginTokens_RememberMe() {
	userID := int64(123)

	accessToken, refreshToken, err := suite.middleware.GenerateLoginTokens(userID, true)

	suite.NoError(err)
	suite.NotEmpty(accessToken)
	suite.NotEmpty(refreshToken)

	// Verify refresh token expiration (1 minute for rememberMe=true)
	refreshClaims, err := suite.decodeToken(refreshToken, "refresh")
	suite.NoError(err)
	suite.WithinDuration(time.Now().Add(1*time.Minute), refreshClaims.ExpiresAt.Time, 2*time.Second)
}

func (suite *AuthMiddlewareTestSuite) TestWebClientAuthentication_ValidAccessToken() {
	userID := int64(123)

	// Generate valid tokens
	accessToken, _, err := suite.middleware.GenerateLoginTokens(userID, false)
	suite.NoError(err)

	handler := suite.middleware.WebClientAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedUserID, ok := middleware.GetUserID(r)
		suite.True(ok)
		suite.Equal(userID, extractedUserID)
		w.WriteHeader(http.StatusOK)
		err := json.NewEncoder(w).Encode(map[string]int64{"user_id": extractedUserID})
		if err != nil {
			fmt.Println("encode error", err.Error())
		}
	}))

	// Make request with valid access token
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access", Value: accessToken})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
}

func (suite *AuthMiddlewareTestSuite) TestWebClientAuthentication_AccessTokenRotation() {
	userID := int64(123)

	// Create expired access token and valid refresh token
	expiredAccessToken := suite.createTokenWithExpiry(userID, "access", -5*time.Second)
	validRefreshToken := suite.createTokenWithExpiry(userID, "refresh", 10*time.Minute)

	handler := suite.middleware.WebClientAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		extractedUserID, ok := middleware.GetUserID(r)
		suite.True(ok)
		suite.Equal(userID, extractedUserID)
		w.WriteHeader(http.StatusOK)
	}))

	// Make request with expired access but valid refresh
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access", Value: expiredAccessToken})
	req.AddCookie(&http.Cookie{Name: "refresh", Value: validRefreshToken})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)
	suite.Equal(http.StatusOK, w.Code)

	// Verify new access token was issued
	setCookieHeaders := w.Header()["Set-Cookie"]
	suite.NotEmpty(setCookieHeaders, "Should have Set-Cookie headers")

	var foundAccessCookie bool
	for _, header := range setCookieHeaders {
		if strings.Contains(header, "access=") {
			foundAccessCookie = true
			suite.NotContains(header, expiredAccessToken, "Should be a new access token")
			break
		}
	}
	suite.True(foundAccessCookie, "New access cookie should be issued")
}

func (suite *AuthMiddlewareTestSuite) TestWebClientAuthentication_BothTokensExpired() {
	userID := int64(123)

	// Create both tokens as expired
	expiredAccessToken := suite.createTokenWithExpiry(userID, "access", -5*time.Second)
	expiredRefreshToken := suite.createTokenWithExpiry(userID, "refresh", -1*time.Second)

	handler := suite.middleware.WebClientAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("Should not reach handler")
	}))

	// Make request with both expired tokens
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access", Value: expiredAccessToken})
	req.AddCookie(&http.Cookie{Name: "refresh", Value: expiredRefreshToken})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return unauthorized
	suite.Equal(http.StatusUnauthorized, w.Code)

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	suite.NoError(err)
	suite.Equal("Unauthorized", response["title"])
	suite.Equal("unauthenticated", response["message"])
}

func (suite *AuthMiddlewareTestSuite) TestWebClientAuthentication_NoTokens() {

	handler := suite.middleware.WebClientAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		suite.Fail("Should not reach handler")
	}))

	// Make request with no tokens
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// Should return unauthorized
	suite.Equal(http.StatusUnauthorized, w.Code)
}

func (suite *AuthMiddlewareTestSuite) createTokenWithExpiry(userID int64, tokenType string, expiryOffset time.Duration) string {
	var jwtKey []byte
	switch tokenType {
	case "access":
		jwtKey = []byte(suite.cfg.Auth.AccessSecret)
	case "refresh":
		jwtKey = []byte(suite.cfg.Auth.RefreshSecret)
	default:
		suite.T().Fatalf("unsupported token type: %s", tokenType)
	}

	encryptedUserID, err := suite.middleware.EncodeWebClientUserID(userID)
	if err != nil {
		suite.T().Fatal(err)
	}

	claims := &middleware.WebClientUserClaim{
		UserID: encryptedUserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiryOffset)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "go-web-template",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		suite.T().Fatal(err)
	}

	return tokenString
}

func (suite *AuthMiddlewareTestSuite) TestEncodeDecodeUserID() {
	userID := int64(12345)

	encoded, err := suite.middleware.EncodeWebClientUserID(userID)
	suite.NoError(err)
	suite.NotEmpty(encoded)

	// The encoded string should be different each time
	encoded2, err := suite.middleware.EncodeWebClientUserID(userID)
	suite.NoError(err)
	suite.NotEqual(encoded, encoded2, "Encryption should use random nonce")
}
