package services_test

import (
	"context"
	"fmt"
	"go-web-template/internal/database"
	"go-web-template/tests/integration"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UserServiceTestSuite struct {
	integration.ServiceIntegrationSuite
}

func TestUserServiceTestSuite(t *testing.T) {
	suite.Run(t, new(UserServiceTestSuite))
}

func (s *UserServiceTestSuite) TestListUsers() {
	svc := s.TC.Services.UserService

	s.seedUser("test1@example.com", "Test User 1")
	s.seedUser("test2@example.com", "Test User 2")

	result, err := svc.ListUsers(context.Background(), 1, 10)

	s.Require().NoError(err)
	s.Assert().Len(result.Data, 2)
	s.Assert().Equal(2, result.Total)
}

func (s *UserServiceTestSuite) TestListUsers_Pagination() {
	svc := s.TC.Services.UserService

	for i := 1; i <= 5; i++ {
		s.seedUser(fmt.Sprintf("user%d@example.com", i), fmt.Sprintf("User %d", i))
	}

	result, err := svc.ListUsers(context.Background(), 1, 2)

	s.Require().NoError(err)
	s.Assert().Len(result.Data, 2)
	s.Assert().Equal(5, result.Total)
	s.Assert().Equal(3, result.TotalPages)
}

func (s *UserServiceTestSuite) seedUser(email, displayName string) {
	role, err := s.TC.Queries.GetRoleByName(context.Background(), "user")
	s.Require().NoError(err)

	_, err = s.TC.Queries.CreateUser(context.Background(), database.CreateUserParams{
		Email:       email,
		Password:    "hashed",
		DisplayName: displayName,
		RoleID:      role.ID,
	})
	s.Require().NoError(err)
}
