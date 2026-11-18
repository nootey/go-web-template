package user_test

import (
	"encoding/json"
	"go-web-template/internal/apperr"
	"go-web-template/internal/domains/user"
	"go-web-template/internal/utils"
	"go-web-template/mocks"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

type UserHandlerTestSuite struct {
	suite.Suite
	router      *chi.Mux
	mockService *mocks.MockUserServiceInterface
	handler     *user.UserHandler
}

func (suite *UserHandlerTestSuite) SetupTest() {
	suite.mockService = mocks.NewMockUserServiceInterface(suite.T())
	suite.handler = user.NewUserHandler(suite.mockService, apperr.NewResponder(zap.NewNop()))

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
	next := "next"
	mockUsers := &utils.Page[*user.User]{
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
		NextCursor: &next,
	}

	suite.mockService.EXPECT().
		ListUsers(mock.Anything, "abc", 10).
		Return(mockUsers, nil).
		Once()

	req := httptest.NewRequest(http.MethodGet, "/users?cursor=abc&page_size=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	suite.Equal(http.StatusOK, w.Code)

	var response utils.Page[*user.User]
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Len(response.Data, 2)
	suite.Require().NotNil(response.NextCursor)
	suite.Equal("next", *response.NextCursor)
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

	var response apperr.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.Require().NoError(err)
	suite.Equal("something went wrong", response.Message)
}

// Test default pagination values
func (suite *UserHandlerTestSuite) TestListUsers_DefaultPagination() {
	suite.mockService.EXPECT().
		ListUsers(mock.Anything, "", 0).
		Return(&utils.Page[*user.User]{Data: []*user.User{}}, nil).
		Twice()

	for _, query := range []string{"", "?page=2&page_size=-x"} {
		req := httptest.NewRequest(http.MethodGet, "/users"+query, nil)
		w := httptest.NewRecorder()

		suite.router.ServeHTTP(w, req)

		suite.Equal(http.StatusOK, w.Code, query)
		suite.JSONEq(`{"data":[],"next_cursor":null}`, w.Body.String(), query)
	}
}
