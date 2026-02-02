package user

import (
	"context"
	"go-web-template/internal/database"
)

type UserServiceInterface interface {
	ListUsers(ctx context.Context, page, pageSize int) (*PaginatedUsers, error)
}

var _ UserServiceInterface = (*UserService)(nil)

type UserService struct {
	queries *database.Queries
}

func NewUserService(queries *database.Queries) *UserService {
	return &UserService{
		queries: queries,
	}
}

func toUserModel(dbUser database.User) *User {
	user := &User{
		ID:          dbUser.ID,
		Email:       dbUser.Email,
		DisplayName: dbUser.DisplayName,
		RoleID:      dbUser.RoleID,
		CreatedAt:   dbUser.CreatedAt,
		UpdatedAt:   dbUser.UpdatedAt,
	}

	if dbUser.EmailConfirmed.Valid {
		user.EmailConfirmed = &dbUser.EmailConfirmed.Time
	}
	if dbUser.DeletedAt.Valid {
		user.DeletedAt = &dbUser.DeletedAt.Time
	}

	return user
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int) (*PaginatedUsers, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	dbUsers, err := s.queries.ListUsersPaginated(ctx, database.ListUsersPaginatedParams{
		Limit:  int32(pageSize),
		Offset: int32(offset),
	})
	if err != nil {
		return nil, err
	}

	total, err := s.queries.CountUsers(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to domain models
	users := make([]*User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toUserModel(dbUser)
	}

	totalPages := (int(total) + pageSize - 1) / pageSize

	return &PaginatedUsers{
		Data:       users,
		Page:       page,
		PageSize:   pageSize,
		Total:      int(total),
		TotalPages: totalPages,
	}, nil
}
