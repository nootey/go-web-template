package auth

import (
	"context"
	"database/sql"
	"errors"
	"go-web-template/internal/database"
	"go-web-template/internal/domains/user"

	"golang.org/x/crypto/bcrypt"
)

type AuthServiceInterface interface {
	ValidateCredentials(ctx context.Context, email, password string) (*user.User, error)
	CreateUser(ctx context.Context, displayName, email, password string) (*user.User, error)
	GetUserByID(ctx context.Context, userID int64) (*user.User, error)
}

var _ AuthServiceInterface = (*AuthService)(nil)

type AuthService struct {
	queries *database.Queries
}

func NewAuthService(queries *database.Queries) *AuthService {
	return &AuthService{
		queries: queries,
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
