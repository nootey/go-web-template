//go:build integration

package auth_test

import (
	"context"
	"go-web-template/config"
	"go-web-template/internal/apperr"
	"go-web-template/internal/domains/auth"
	"go-web-template/tests/integration"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	integration.ServiceIntegrationSuite
}

func TestIntegrationAuthServiceTestSuite(t *testing.T) {
	suite.Run(t, new(AuthServiceTestSuite))
}

func (s *AuthServiceTestSuite) tokenFrom(link string) string {
	u, err := url.Parse(link)
	s.Require().NoError(err)
	token := u.Query().Get("token")
	s.Require().NotEmpty(token, "expected a token in link %q", link)
	return token
}

func (s *AuthServiceTestSuite) TestRegister_UnconfirmedUserCannotLogin() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	err := svc.Register(ctx, "Jane", "jane@example.com", "supersecret")
	s.Require().NoError(err)

	link, err := url.Parse(s.TC.Mailer.ConfirmationLink)
	s.Require().NoError(err)
	s.Equal("localhost:5173", link.Host)

	// Unconfirmed account: login is blocked.
	_, err = svc.ValidateCredentials(ctx, "jane@example.com", "supersecret")
	s.ErrorIs(err, auth.ErrEmailNotConfirmed)
}

func (s *AuthServiceTestSuite) TestRegister_SignupOff() {
	svc := auth.NewAuthService(s.TC.Queries, nil, s.TC.Mailer, &config.Config{})
	ctx := context.Background()

	s.False(svc.SignupAllowed())
	err := svc.Register(ctx, "Jane", "jane@example.com", "supersecret")

	var appErr *apperr.Error
	s.Require().ErrorAs(err, &appErr)
	s.Equal(apperr.KindForbidden, appErr.Kind)
	s.Empty(s.TC.Mailer.ConfirmationLink)

	var users int
	s.Require().NoError(s.TC.DB.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&users))
	s.Zero(users)
}

func (s *AuthServiceTestSuite) TestConfirmEmail_UnlocksLogin() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "supersecret"))

	token := s.tokenFrom(s.TC.Mailer.ConfirmationLink)
	s.Require().NoError(svc.ConfirmEmail(ctx, token))

	u, err := svc.ValidateCredentials(ctx, "jane@example.com", "supersecret")
	s.Require().NoError(err)
	s.Equal("jane@example.com", u.Email)
}

func (s *AuthServiceTestSuite) TestConfirmEmail_TokenIsSingleUse() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "supersecret"))
	token := s.tokenFrom(s.TC.Mailer.ConfirmationLink)

	s.Require().NoError(svc.ConfirmEmail(ctx, token))
	s.Error(svc.ConfirmEmail(ctx, token))
}

func (s *AuthServiceTestSuite) TestResendConfirmation() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "supersecret"))
	s.TC.Mailer.ConfirmationLink = "" // clear the register email

	// Unconfirmed user: a fresh link is issued.
	s.Require().NoError(svc.ResendConfirmation(ctx, "jane@example.com"))
	s.Require().NotEmpty(s.TC.Mailer.ConfirmationLink)
	confirmToken := s.tokenFrom(s.TC.Mailer.ConfirmationLink)

	// Unknown email: swallowed, no link.
	s.TC.Mailer.ConfirmationLink = ""
	s.Require().NoError(svc.ResendConfirmation(ctx, "nobody@example.com"))
	s.Empty(s.TC.Mailer.ConfirmationLink)

	// Once confirmed, resend is swallowed (no new link).
	s.Require().NoError(svc.ConfirmEmail(ctx, confirmToken))
	s.TC.Mailer.ConfirmationLink = ""
	s.Require().NoError(svc.ResendConfirmation(ctx, "jane@example.com"))
	s.Empty(s.TC.Mailer.ConfirmationLink)
}

func (s *AuthServiceTestSuite) TestRequestPasswordReset_UnknownEmailSwallowed() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.RequestPasswordReset(ctx, "nobody@example.com"))
	s.Empty(s.TC.Mailer.ResetLink)
}

func (s *AuthServiceTestSuite) TestResetPassword_ChangesPasswordAndReturnsUserID() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	// Register + confirm so the account can log in.
	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "oldsecret1"))
	s.Require().NoError(svc.ConfirmEmail(ctx, s.tokenFrom(s.TC.Mailer.ConfirmationLink)))

	confirmed, err := svc.ValidateCredentials(ctx, "jane@example.com", "oldsecret1")
	s.Require().NoError(err)

	s.Require().NoError(svc.RequestPasswordReset(ctx, "jane@example.com"))
	resetToken := s.tokenFrom(s.TC.Mailer.ResetLink)

	userID, err := svc.ResetPassword(ctx, resetToken, "newsecret1", "newsecret1")
	s.Require().NoError(err)
	s.Equal(confirmed.ID, userID)

	// Old password no longer works, new one does.
	_, err = svc.ValidateCredentials(ctx, "jane@example.com", "oldsecret1")
	s.Require().Error(err)

	u, err := svc.ValidateCredentials(ctx, "jane@example.com", "newsecret1")
	s.Require().NoError(err)
	s.Equal(confirmed.ID, u.ID)
}

func (s *AuthServiceTestSuite) TestResetPassword_Rejections() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "oldsecret1"))
	s.Require().NoError(svc.ConfirmEmail(ctx, s.tokenFrom(s.TC.Mailer.ConfirmationLink)))
	s.Require().NoError(svc.RequestPasswordReset(ctx, "jane@example.com"))
	resetToken := s.tokenFrom(s.TC.Mailer.ResetLink)

	// Mismatched confirmation is rejected without consuming the token.
	_, err := svc.ResetPassword(ctx, resetToken, "newsecret1", "different1")
	s.Require().Error(err)

	// Too-weak password is rejected.
	_, err = svc.ResetPassword(ctx, resetToken, "short", "short")
	s.Require().Error(err)

	// Unknown token is rejected.
	_, err = svc.ResetPassword(ctx, "bogus-token", "newsecret1", "newsecret1")
	s.Require().Error(err)
}

func (s *AuthServiceTestSuite) TestValidateCredentials_WrongPassword() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "supersecret"))
	s.Require().NoError(svc.ConfirmEmail(ctx, s.tokenFrom(s.TC.Mailer.ConfirmationLink)))

	_, err := svc.ValidateCredentials(ctx, "jane@example.com", "wrongpassword")
	s.Require().Error(err)
	s.Equal("invalid credentials", err.Error())
}
