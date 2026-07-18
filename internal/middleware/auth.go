package middleware

import (
	"context"
	"encoding/json"
	"net/http"

	"go-web-template/internal/config"
	"go-web-template/internal/sessions"

	"go.uber.org/zap"
)

// Context key for user ID
type ctxKey int

const userIDKey ctxKey = 0

type AuthMiddlewareInterface interface {
	WebClientAuthentication(next http.Handler) http.Handler
	CreateLoginSession(ctx context.Context, userID int64, rememberMe bool) (id string, maxAge int, err error)
	DestroySession(ctx context.Context, sessionID string) error
	RevokeAllSessions(ctx context.Context, userID int64) error
	SetSessionCookie(w http.ResponseWriter, id string, maxAge int)
	ClearSessionCookie(w http.ResponseWriter)
}

type AuthMiddleware struct {
	store  *sessions.Store
	cfg    *config.Config
	logger *zap.Logger
}

func NewAuthMiddleware(store *sessions.Store, cfg *config.Config, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		store:  store,
		cfg:    cfg,
		logger: logger,
	}
}

var _ AuthMiddlewareInterface = (*AuthMiddleware)(nil)

func (m *AuthMiddleware) WebClientAuthentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessions.CookieName)
		if err != nil || cookie.Value == "" {
			m.respondUnauthorized(w, "unauthenticated")
			return
		}

		userID, err := m.store.Validate(r.Context(), cookie.Value)
		if err != nil {
			m.respondUnauthorized(w, "unauthenticated")
			return
		}

		// Store user ID in context and continue
		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CreateLoginSession stores a new session and returns its id alongside the
// cookie Max-Age, which matches the session's Redis TTL.
func (m *AuthMiddleware) CreateLoginSession(ctx context.Context, userID int64, rememberMe bool) (string, int, error) {
	id, err := m.store.Create(ctx, userID, rememberMe)
	if err != nil {
		return "", 0, err
	}
	return id, int(m.store.TTL(rememberMe).Seconds()), nil
}

func (m *AuthMiddleware) DestroySession(ctx context.Context, sessionID string) error {
	return m.store.Delete(ctx, sessionID)
}

func (m *AuthMiddleware) RevokeAllSessions(ctx context.Context, userID int64) error {
	return m.store.DeleteAllForUser(ctx, userID)
}

func (m *AuthMiddleware) SetSessionCookie(w http.ResponseWriter, id string, maxAge int) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessions.CookieName,
		Value:    id,
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   maxAge,
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (m *AuthMiddleware) ClearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessions.CookieName,
		Value:    "",
		Path:     "/",
		Domain:   m.cfg.App.CookieDomain,
		MaxAge:   -1,
		Secure:   m.cfg.App.Environment == "production",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
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
