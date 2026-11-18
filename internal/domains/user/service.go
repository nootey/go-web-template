package user

import (
	"context"
	"database/sql"
	"go-web-template/internal/database"
	"go-web-template/internal/utils"
)

type UserServiceInterface interface {
	ListUsers(ctx context.Context, cursor string, pageSize int) (*utils.Page[*User], error)
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

func (s *UserService) ListUsers(ctx context.Context, cursor string, pageSize int) (*utils.Page[*User], error) {
	var after userCursor
	if cursor != "" {
		if err := utils.DecodeCursor(cursor, &after); err != nil {
			return nil, err
		}
	}
	size := utils.PageSize(pageSize)

	dbUsers, err := s.queries.ListUsersPage(ctx, database.ListUsersPageParams{
		AfterCreatedAt: sql.NullTime{Time: after.CreatedAt, Valid: cursor != ""},
		AfterID:        sql.NullInt64{Int64: after.ID, Valid: cursor != ""},
		Limit:          size + 1,
	})
	if err != nil {
		return nil, err
	}

	users := make([]*User, len(dbUsers))
	for i, dbUser := range dbUsers {
		users[i] = toUserModel(dbUser)
	}

	return utils.NewPage(users, size, func(u *User) any {
		return userCursor{CreatedAt: u.CreatedAt, ID: u.ID}
	})
}
