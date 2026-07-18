package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-web-template/internal/domains/auth"
	"go-web-template/internal/domains/user"
	"go-web-template/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type AuthHandlerTestSuite struct {
	suite.Suite
	router         *chi.Mux
	mockService    *mocks.MockAuthServiceInterface
	mockMiddleware *mocks.MockAuthMiddlewareInterface
	handler        *auth.AuthHandler
}

func TestAuthHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(AuthHandlerTestSuite))
}

func (suite *AuthHandlerTestSuite) SetupTest() {
	suite.mockService = mocks.NewMockAuthServiceInterface(suite.T())
	suite.mockMiddleware = mocks.NewMockAuthMiddlewareInterface(suite.T())
	suite.handler = auth.NewAuthHandler(suite.mockService, suite.mockMiddleware, zap.NewNop())

	// chi invokes this at mount time for the protected group; stub as passthrough.
	suite.mockMiddleware.EXPECT().
		WebClientAuthentication(mock.Anything).
		RunAndReturn(func(next http.Handler) http.Handler { return next }).
		Maybe()

	suite.router = chi.NewRouter()
	suite.router.Mount("/auth", suite.handler.Routes())
}

func (suite *AuthHandlerTestSuite) do(method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		suite.Require().NoError(json.NewEncoder(&buf).Encode(body))
	}
	req := httptest.NewRequest(method, path, &buf)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	return w
}

func (suite *AuthHandlerTestSuite) TestRegister_Success() {
	suite.mockService.EXPECT().
		Register(mock.Anything, "Jane", "jane@example.com", "supersecret").
		Return(nil).
		Once()

	w := suite.do(http.MethodPost, "/auth/register", auth.RegisterRequest{
		DisplayName:          "Jane",
		Email:                "jane@example.com",
		Password:             "supersecret",
		PasswordConfirmation: "supersecret",
	})

	suite.Equal(http.StatusCreated, w.Code)
	suite.Empty(w.Result().Cookies()) // no auto-login
}

func (suite *AuthHandlerTestSuite) TestRegister_PasswordMismatch() {
	w := suite.do(http.MethodPost, "/auth/register", auth.RegisterRequest{
		DisplayName:          "Jane",
		Email:                "jane@example.com",
		Password:             "supersecret",
		PasswordConfirmation: "different",
	})

	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLogin_EmailNotConfirmed() {
	suite.mockService.EXPECT().
		ValidateCredentials(mock.Anything, "jane@example.com", "supersecret").
		Return(nil, auth.ErrEmailNotConfirmed).
		Once()

	w := suite.do(http.MethodPost, "/auth/login", auth.LoginRequest{
		Email:    "jane@example.com",
		Password: "supersecret",
	})

	suite.Equal(http.StatusForbidden, w.Code)
}

func (suite *AuthHandlerTestSuite) TestLogin_Success() {
	suite.mockService.EXPECT().
		ValidateCredentials(mock.Anything, "jane@example.com", "supersecret").
		Return(&user.User{ID: 7, Email: "jane@example.com", DisplayName: "Jane"}, nil).
		Once()
	suite.mockMiddleware.EXPECT().
		CreateLoginSession(mock.Anything, int64(7), false).
		Return("sess-id", 3600, nil).
		Once()
	suite.mockMiddleware.EXPECT().
		SetSessionCookie(mock.Anything, "sess-id", 3600).
		Return().
		Once()

	w := suite.do(http.MethodPost, "/auth/login", auth.LoginRequest{
		Email:    "jane@example.com",
		Password: "supersecret",
	})

	suite.Equal(http.StatusOK, w.Code)
}

func (suite *AuthHandlerTestSuite) TestConfirmEmail_Success() {
	suite.mockService.EXPECT().
		ConfirmEmail(mock.Anything, "tok123").
		Return(nil).
		Once()

	w := suite.do(http.MethodGet, "/auth/confirm-email?token=tok123", nil)

	suite.Equal(http.StatusOK, w.Code)
}

func (suite *AuthHandlerTestSuite) TestConfirmEmail_MissingToken() {
	w := suite.do(http.MethodGet, "/auth/confirm-email", nil)
	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestConfirmEmail_InvalidToken() {
	suite.mockService.EXPECT().
		ConfirmEmail(mock.Anything, "bad").
		Return(assert.AnError).
		Once()

	w := suite.do(http.MethodGet, "/auth/confirm-email?token=bad", nil)
	suite.Equal(http.StatusBadRequest, w.Code)
}

func (suite *AuthHandlerTestSuite) TestResendConfirmation_Generic() {
	suite.mockService.EXPECT().
		ResendConfirmation(mock.Anything, "jane@example.com").
		Return(nil).
		Once()

	w := suite.do(http.MethodPost, "/auth/resend-confirmation", auth.ConfirmResendRequest{Email: "jane@example.com"})
	suite.Equal(http.StatusOK, w.Code)
}

func (suite *AuthHandlerTestSuite) TestRequestPasswordReset_Generic() {
	suite.mockService.EXPECT().
		RequestPasswordReset(mock.Anything, "jane@example.com").
		Return(nil).
		Once()

	w := suite.do(http.MethodPost, "/auth/request-password-reset", auth.RequestResetRequest{Email: "jane@example.com"})
	suite.Equal(http.StatusOK, w.Code)
}

func (suite *AuthHandlerTestSuite) TestResetPassword_Success() {
	suite.mockService.EXPECT().
		ResetPassword(mock.Anything, "tok", "newsecret1", "newsecret1").
		Return(int64(7), nil).
		Once()
	suite.mockMiddleware.EXPECT().
		RevokeAllSessions(mock.Anything, int64(7)).
		Return(nil).
		Once()

	w := suite.do(http.MethodPost, "/auth/reset-password", auth.ResetPasswordRequest{
		Token:                "tok",
		Password:             "newsecret1",
		PasswordConfirmation: "newsecret1",
	})

	suite.Equal(http.StatusOK, w.Code)
}

func (suite *AuthHandlerTestSuite) TestResetPassword_InvalidToken() {
	suite.mockService.EXPECT().
		ResetPassword(mock.Anything, "bad", "newsecret1", "newsecret1").
		Return(int64(0), assert.AnError).
		Once()

	w := suite.do(http.MethodPost, "/auth/reset-password", auth.ResetPasswordRequest{
		Token:                "bad",
		Password:             "newsecret1",
		PasswordConfirmation: "newsecret1",
	})

	suite.Equal(http.StatusBadRequest, w.Code)
}
