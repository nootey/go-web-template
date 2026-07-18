package auth_test

import (
	"context"
	"net/url"
	"testing"

	"go-web-template/internal/domains/auth"
	"go-web-template/tests/integration"

	"github.com/stretchr/testify/suite"
)

type AuthServiceTestSuite struct {
	integration.ServiceIntegrationSuite
}

func TestAuthServiceTestSuite(t *testing.T) {
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

	s.NotEmpty(s.TC.Mailer.ConfirmationLink)

	// Unconfirmed account: login is blocked.
	_, err = svc.ValidateCredentials(ctx, "jane@example.com", "supersecret")
	s.ErrorIs(err, auth.ErrEmailNotConfirmed)
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
	s.Error(err)

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
	s.Error(err)

	// Too-weak password is rejected.
	_, err = svc.ResetPassword(ctx, resetToken, "short", "short")
	s.Error(err)

	// Unknown token is rejected.
	_, err = svc.ResetPassword(ctx, "bogus-token", "newsecret1", "newsecret1")
	s.Error(err)
}

func (s *AuthServiceTestSuite) TestValidateCredentials_WrongPassword() {
	svc := s.TC.Services.AuthService
	ctx := context.Background()

	s.Require().NoError(svc.Register(ctx, "Jane", "jane@example.com", "supersecret"))
	s.Require().NoError(svc.ConfirmEmail(ctx, s.tokenFrom(s.TC.Mailer.ConfirmationLink)))

	_, err := svc.ValidateCredentials(ctx, "jane@example.com", "wrongpassword")
	s.Error(err)
	s.Equal("invalid credentials", err.Error())
}
