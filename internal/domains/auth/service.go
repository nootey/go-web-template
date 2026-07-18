package auth

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"go-web-template/internal/config"
	"go-web-template/internal/database"
	"go-web-template/internal/domains/user"
	"go-web-template/internal/mailer"
	"go-web-template/internal/tokens"
	"go-web-template/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

// ErrEmailNotConfirmed means the credentials are valid but the email is unconfirmed.
var ErrEmailNotConfirmed = errors.New("please confirm your email address before logging in")

type AuthServiceInterface interface {
	ValidateCredentials(ctx context.Context, email, password string) (*user.User, error)
	CreateUser(ctx context.Context, displayName, email, password string) (*user.User, error)
	GetUserByID(ctx context.Context, userID int64) (*user.User, error)
	Register(ctx context.Context, displayName, email, password string) error
	ConfirmEmail(ctx context.Context, tokenID string) error
	ResendConfirmation(ctx context.Context, email string) error
	RequestPasswordReset(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, tokenID, password, confirmation string) (int64, error)
}

var _ AuthServiceInterface = (*AuthService)(nil)

type AuthService struct {
	queries *database.Queries
	tokens  tokens.TokenStoreInterface
	mailer  mailer.MailerInterface
	cfg     *config.Config
}

func NewAuthService(
	queries *database.Queries,
	tokenStore tokens.TokenStoreInterface,
	mail mailer.MailerInterface,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		queries: queries,
		tokens:  tokenStore,
		mailer:  mail,
		cfg:     cfg,
	}
}

func (s *AuthService) ValidateCredentials(ctx context.Context, email, password string) (*user.User, error) {
	dbUser, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Checked after the password so an attacker can't probe which emails exist.
	if !dbUser.EmailConfirmed.Valid {
		return nil, ErrEmailNotConfirmed
	}

	return &user.User{
		ID:          dbUser.ID,
		Email:       dbUser.Email,
		DisplayName: dbUser.DisplayName,
		RoleID:      dbUser.RoleID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
	}, nil
}

func (s *AuthService) CreateUser(ctx context.Context, displayName, email, password string) (*user.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	defaultRole, err := s.queries.GetDefaultRole(ctx)
	if err != nil {
		return nil, errors.New("failed to create user")
	}

	dbUser, err := s.queries.CreateUser(ctx, database.CreateUserParams{
		Email:       email,
		Password:    string(hashedPassword),
		DisplayName: displayName,
		RoleID:      defaultRole.ID,
	})
	if err != nil {
		return nil, err
	}

	return &user.User{
		ID:          dbUser.ID,
		Email:       dbUser.Email,
		DisplayName: dbUser.DisplayName,
		RoleID:      dbUser.RoleID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
	}, nil
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*user.User, error) {
	dbUser, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user.User{
		ID:          dbUser.ID,
		Email:       dbUser.Email,
		DisplayName: dbUser.DisplayName,
		RoleID:      dbUser.RoleID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
	}, nil
}

// Register creates an unconfirmed user and emails a confirmation link. No session
// is created: the user must confirm before logging in.
func (s *AuthService) Register(ctx context.Context, displayName, email, password string) error {
	if err := validatePasswordStrength(password); err != nil {
		return err
	}

	u, err := s.CreateUser(ctx, displayName, email, password)
	if err != nil {
		return err
	}

	return s.issueConfirmation(ctx, u.ID, u.Email, u.DisplayName)
}

func (s *AuthService) ConfirmEmail(ctx context.Context, tokenID string) error {
	userID, err := s.tokens.Consume(ctx, tokens.PurposeConfirm, tokenID)
	if err != nil {
		return err
	}
	return s.queries.ConfirmUserEmail(ctx, userID)
}

// ResendConfirmation swallows not-found/already-confirmed so the handler can stay generic.
func (s *AuthService) ResendConfirmation(ctx context.Context, email string) error {
	dbUser, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}
	if dbUser.EmailConfirmed.Valid {
		return nil
	}
	return s.issueConfirmation(ctx, dbUser.ID, dbUser.Email, dbUser.DisplayName)
}

// RequestPasswordReset swallows not-found so the handler can stay generic.
func (s *AuthService) RequestPasswordReset(ctx context.Context, email string) error {
	dbUser, err := s.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		return err
	}

	ttl := time.Duration(s.cfg.Token.ResetTTLMinutes) * time.Minute
	id, err := s.tokens.Create(ctx, tokens.PurposeReset, dbUser.ID, ttl)
	if err != nil {
		return err
	}

	link := utils.WebClientBaseURL(s.cfg) + "/reset-password?token=" + id
	return s.mailer.SendPasswordResetEmail(dbUser.Email, dbUser.DisplayName, link)
}

// ResetPassword returns the user id so the caller can revoke that user's sessions.
func (s *AuthService) ResetPassword(ctx context.Context, tokenID, password, confirmation string) (int64, error) {
	if password != confirmation {
		return 0, errors.New("passwords do not match")
	}
	if err := validatePasswordStrength(password); err != nil {
		return 0, err
	}

	userID, err := s.tokens.Consume(ctx, tokens.PurposeReset, tokenID)
	if err != nil {
		return 0, err
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}

	if err := s.queries.UpdateUserPassword(ctx, database.UpdateUserPasswordParams{
		ID:       userID,
		Password: string(hashed),
	}); err != nil {
		return 0, err
	}

	return userID, nil
}

func (s *AuthService) issueConfirmation(ctx context.Context, userID int64, email, displayName string) error {
	ttl := time.Duration(s.cfg.Token.ConfirmTTLHours) * time.Hour
	id, err := s.tokens.Create(ctx, tokens.PurposeConfirm, userID, ttl)
	if err != nil {
		return err
	}

	link := utils.APIBaseURL(s.cfg) + "/api/auth/confirm-email?token=" + id
	return s.mailer.SendConfirmationEmail(email, displayName, link)
}

func validatePasswordStrength(pw string) error {
	if len(pw) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	return nil
}
