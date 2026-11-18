//go:build integration

package user_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"go-web-template/internal/apperr"
	"go-web-template/internal/database"
	"go-web-template/internal/domains/user"
	"go-web-template/tests/integration"
	"net/http"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	integration.ServiceIntegrationSuite
}

func TestIntegrationUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestListUsers() {
	svc := s.TC.Services.UserService

	s.seedUser("test1@example.com", "Test User 1")
	s.seedUser("test2@example.com", "Test User 2")
	deleted := s.seedUser("test3@example.com", "Test User 3")
	_, err := s.TC.DB.ExecContext(context.Background(), `UPDATE users SET deleted_at = NOW() WHERE id = $1`, deleted.ID)
	s.Require().NoError(err)

	result, err := svc.ListUsers(context.Background(), "", 10)

	s.Require().NoError(err)
	s.ElementsMatch([]string{"test1@example.com", "test2@example.com"}, userEmails(result.Data))
	s.Nil(result.NextCursor)
}

// Newest first. Users 2 to 5 tie on created_at, so a cursor inside the tie must keep the id too.
func (s *UserServiceTestSuite) TestListUsers_Pages() {
	for i := 1; i <= 5; i++ {
		s.seedUser(fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i))
	}
	_, err := s.TC.DB.ExecContext(context.Background(), `
		UPDATE users SET created_at = CASE WHEN email = 'user1@example.com'
			THEN TIMESTAMP '2026-03-01 10:00:00' ELSE TIMESTAMP '2026-01-01 10:00:00' END`)
	s.Require().NoError(err)

	pages := s.walkPages(2)

	s.Equal([][]string{
		{"user1@example.com", "user5@example.com"},
		{"user4@example.com", "user3@example.com"},
		{"user2@example.com"},
	}, pages)
}

func (s *UserServiceTestSuite) TestListUsers_BadCursor() {
	svc := s.TC.Services.UserService
	s.seedUser("test1@example.com", "Test User 1")

	cases := []struct {
		label  string
		cursor string
	}{
		{"not base64", "!!!"},
		{"not json", cursorOf("nope")},
		{"date not a time", cursorOf(`{"c":"not a time","i":1}`)},
		{"id not a number", cursorOf(`{"c":"2026-01-01T10:00:00Z","i":"abc"}`)},
	}
	for _, tc := range cases {
		_, err := svc.ListUsers(context.Background(), tc.cursor, 10)

		status, msg := apperr.Resolve(err)
		s.Equal(http.StatusBadRequest, status, tc.label)
		s.Equal("invalid cursor", msg, tc.label)
	}
}

func (s *UserServiceTestSuite) seedUser(email, displayName string) database.User {
	role, err := s.TC.Queries.GetRoleByName(context.Background(), "user")
	s.Require().NoError(err)

	u, err := s.TC.Queries.CreateUser(context.Background(), database.CreateUserParams{
		Email:       email,
		Password:    "hashed",
		DisplayName: displayName,
		RoleID:      role.ID,
	})
	s.Require().NoError(err)
	return u
}

func (s *UserServiceTestSuite) walkPages(pageSize int) [][]string {
	var pages [][]string
	cursor := ""
	for range 50 {
		result, err := s.TC.Services.UserService.ListUsers(context.Background(), cursor, pageSize)
		s.Require().NoError(err)

		pages = append(pages, userEmails(result.Data))
		if result.NextCursor == nil {
			return pages
		}
		cursor = *result.NextCursor
	}
	s.FailNow("the pages never end")
	return nil
}

func cursorOf(raw string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(raw))
}

func userEmails(users []*user.User) []string {
	emails := make([]string, len(users))
	for i, u := range users {
		emails[i] = u.Email
	}
	return emails
}
