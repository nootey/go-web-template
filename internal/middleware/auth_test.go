package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go-web-template/internal/config"
	"go-web-template/internal/middleware"
	"go-web-template/internal/sessions"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	mr         *miniredis.Miniredis
	store      *sessions.Store
	middleware *middleware.AuthMiddleware
}

func TestAuthMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}

func (suite *AuthMiddlewareTestSuite) SetupTest() {
	if err := config.Load(); err != nil {
		panic(err)
	}
	cfg := config.Get()

	suite.mr = miniredis.RunT(suite.T())
	rdb := redis.NewClient(&redis.Options{Addr: suite.mr.Addr()})
	suite.store = sessions.NewStore(rdb, cfg.Session)

	logger, _ := zap.NewDevelopment()
	suite.middleware = middleware.NewAuthMiddleware(suite.store, cfg, logger)
}

// protectedHandler records the user ID the middleware put in the request context.
func (suite *AuthMiddlewareTestSuite) protectedHandler(gotUserID *int64) http.Handler {
	return suite.middleware.WebClientAuthentication(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r)
		suite.True(ok, "expected user ID in context")
		*gotUserID = userID
		w.WriteHeader(http.StatusOK)
	}))
}

func (suite *AuthMiddlewareTestSuite) assertUnauthorized(rec *httptest.ResponseRecorder) {
	suite.Equal(http.StatusUnauthorized, rec.Code)

	var body map[string]string
	suite.Require().NoError(json.NewDecoder(rec.Body).Decode(&body))
	suite.Equal("Unauthorized", body["title"])
	suite.Equal("unauthenticated", body["message"])
}

func (suite *AuthMiddlewareTestSuite) TestValidSession() {
	id, maxAge, err := suite.middleware.CreateLoginSession(suite.T().Context(), 42, false)
	suite.Require().NoError(err)
	suite.Equal(int(24*time.Hour.Seconds()), maxAge)

	var gotUserID int64
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName, Value: id})
	rec := httptest.NewRecorder()

	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)

	suite.Equal(http.StatusOK, rec.Code)
	suite.Equal(int64(42), gotUserID)
}

func (suite *AuthMiddlewareTestSuite) TestMissingCookie() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	var gotUserID int64
	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)

	suite.assertUnauthorized(rec)
}

func (suite *AuthMiddlewareTestSuite) TestEmptyCookie() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName, Value: ""})
	rec := httptest.NewRecorder()

	var gotUserID int64
	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)

	suite.assertUnauthorized(rec)
}

func (suite *AuthMiddlewareTestSuite) TestUnknownSession() {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName, Value: "not-a-real-session"})
	rec := httptest.NewRecorder()

	var gotUserID int64
	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)

	suite.assertUnauthorized(rec)
}

func (suite *AuthMiddlewareTestSuite) TestExpiredSession() {
	id, _, err := suite.middleware.CreateLoginSession(suite.T().Context(), 42, false)
	suite.Require().NoError(err)

	suite.mr.FastForward(24 * time.Hour)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName, Value: id})
	rec := httptest.NewRecorder()

	var gotUserID int64
	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)

	suite.assertUnauthorized(rec)
}

func (suite *AuthMiddlewareTestSuite) TestDestroySessionRevokesAccess() {
	id, _, err := suite.middleware.CreateLoginSession(suite.T().Context(), 42, false)
	suite.Require().NoError(err)

	// Session works before logout.
	var gotUserID int64
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName, Value: id})
	rec := httptest.NewRecorder()
	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)
	suite.Require().Equal(http.StatusOK, rec.Code)

	suite.Require().NoError(suite.middleware.DestroySession(suite.T().Context(), id))
	suite.False(suite.mr.Exists("session:"+id), "expected redis key to be gone after logout")

	// Same cookie is now rejected.
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sessions.CookieName, Value: id})
	rec = httptest.NewRecorder()
	suite.protectedHandler(&gotUserID).ServeHTTP(rec, req)

	suite.assertUnauthorized(rec)
}

func (suite *AuthMiddlewareTestSuite) TestCreateLoginSessionRememberMe() {
	_, maxAge, err := suite.middleware.CreateLoginSession(suite.T().Context(), 42, true)
	suite.Require().NoError(err)

	suite.Equal(int(720*time.Hour.Seconds()), maxAge)
}

func (suite *AuthMiddlewareTestSuite) TestSetSessionCookie() {
	rec := httptest.NewRecorder()
	suite.middleware.SetSessionCookie(rec, "session-id", 3600)

	cookies := rec.Result().Cookies()
	suite.Require().Len(cookies, 1)

	c := cookies[0]
	suite.Equal(sessions.CookieName, c.Name)
	suite.Equal("session-id", c.Value)
	suite.Equal("/", c.Path)
	suite.Equal(3600, c.MaxAge)
	suite.True(c.HttpOnly)
	suite.Equal(http.SameSiteLaxMode, c.SameSite)
	suite.False(c.Secure, "Secure should be off outside production")
}

func (suite *AuthMiddlewareTestSuite) TestClearSessionCookie() {
	rec := httptest.NewRecorder()
	suite.middleware.ClearSessionCookie(rec)

	cookies := rec.Result().Cookies()
	suite.Require().Len(cookies, 1)

	c := cookies[0]
	suite.Equal(sessions.CookieName, c.Name)
	suite.Empty(c.Value)
	suite.Equal(-1, c.MaxAge)
	suite.True(c.HttpOnly)
}
