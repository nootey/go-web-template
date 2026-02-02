package user_test

import (
	"encoding/json"
	"go-web-template/internal/domains/user"
	"go-web-template/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"

	"go-web-template/mocks"
)

type UserHandlerTestSuite struct {
	suite.Suite
	router      *chi.Mux
	mockService *mocks.MockUserServiceInterface
	handler     *user.UserHandler
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.mockService = mocks.NewMockUserServiceInterface(suite.T())
	suite.handler = user.NewUserHandler(suite.mockService, zap.NewNop())

	// Setup router
	suite.router = chi.NewRouter()
	suite.router.Mount("/users", suite.handler.Routes())
}

func (suite *UserHandlerTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

func TestUserHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(UserHandlerTestSuite))
}

// Test successful user listing
func (suite *UserHandlerTestSuite) TestListUsers_Success() {
	mockUsers := &user.PaginatedUsers{
		Data: []*user.User{
			{
				ID:          1,
				Email:       "test1@example.com",
				DisplayName: "Test User 1",
			},
			{
				ID:          2,
				Email:       "test2@example.com",
				DisplayName: "Test User 2",
			},
		},
		Page:       1,
		PageSize:   10,
		Total:      2,
		TotalPages: 1,
	}

	suite.mockService.EXPECT().
		ListUsers(mock.Anything, 1, 10).
		Return(mockUsers, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users?page=1&page_size=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response user.PaginatedUsers
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Len(response.Data, 2)
	suite.Equal(2, response.Total)
}

// Test with pagination parameters
func (suite *UserHandlerTestSuite) TestListUsers_WithPagination() {
	mockUsers := &user.PaginatedUsers{
		Data:       []*user.User{},
		Page:       2,
		PageSize:   5,
		Total:      10,
		TotalPages: 2,
	}

	suite.mockService.EXPECT().
		ListUsers(mock.Anything, 2, 5).
		Return(mockUsers, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users?page=2&page_size=5", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response user.PaginatedUsers
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(2, response.Page)
	suite.Equal(5, response.PageSize)
}

// Test service error handling
func (suite *UserHandlerTestSuite) TestListUsers_ServiceError() {
	suite.mockService.EXPECT().
		ListUsers(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, assert.AnError).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusInternalServerError, w.Code)

	var response utils.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("failed to fetch users", response.Message)
}

// Test default pagination values
func (suite *UserHandlerTestSuite) TestListUsers_DefaultPagination() {
	// When page/page_size not provided, service should receive defaults
	suite.mockService.EXPECT().
		ListUsers(mock.Anything, 0, 0).
		Return(&user.PaginatedUsers{
			Data:       []*user.User{},
			Page:       1,
			PageSize:   20,
			Total:      0,
			TotalPages: 0,
		}, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)
}
