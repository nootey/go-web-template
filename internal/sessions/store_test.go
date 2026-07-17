package sessions_test

import (
	"context"
	"testing"
	"time"

	"go-web-template/internal/config"
	"go-web-template/internal/sessions"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
	mr    *miniredis.Miniredis
	store *sessions.Store
	ctx   context.Context
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (suite *StoreTestSuite) SetupTest() {
	suite.mr = miniredis.RunT(suite.T())
	rdb := redis.NewClient(&redis.Options{Addr: suite.mr.Addr()})
	suite.store = sessions.NewStore(rdb, config.SessionConfig{
		TTLHours:           24,
		RememberMeTTLHours: 720,
	})
	suite.ctx = context.Background()
}

func (suite *StoreTestSuite) TestTTL() {
	suite.Equal(24*time.Hour, suite.store.TTL(false))
	suite.Equal(720*time.Hour, suite.store.TTL(true))
}

func (suite *StoreTestSuite) TestCreateValidateRoundTrip() {
	id, err := suite.store.Create(suite.ctx, 42, false)
	suite.Require().NoError(err)
	suite.NotEmpty(id)

	userID, err := suite.store.Validate(suite.ctx, id)
	suite.Require().NoError(err)
	suite.Equal(int64(42), userID)
}

func (suite *StoreTestSuite) TestCreateGeneratesDistinctIDs() {
	first, err := suite.store.Create(suite.ctx, 1, false)
	suite.Require().NoError(err)
	second, err := suite.store.Create(suite.ctx, 1, false)
	suite.Require().NoError(err)

	suite.NotEqual(first, second)
}

func (suite *StoreTestSuite) TestCreateSetsTTL() {
	tests := []struct {
		name       string
		rememberMe bool
		want       time.Duration
	}{
		{"default", false, 24 * time.Hour},
		{"remember me", true, 720 * time.Hour},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			id, err := suite.store.Create(suite.ctx, 7, tt.rememberMe)
			suite.Require().NoError(err)

			suite.Equal(tt.want, suite.mr.TTL("session:"+id))
		})
	}
}

func (suite *StoreTestSuite) TestValidateUnknownSession() {
	_, err := suite.store.Validate(suite.ctx, "does-not-exist")
	suite.ErrorIs(err, sessions.ErrNotFound)
}

func (suite *StoreTestSuite) TestValidateExpiredSession() {
	id, err := suite.store.Create(suite.ctx, 42, false)
	suite.Require().NoError(err)

	suite.mr.FastForward(24 * time.Hour)

	_, err = suite.store.Validate(suite.ctx, id)
	suite.ErrorIs(err, sessions.ErrNotFound)
}

func (suite *StoreTestSuite) TestDelete() {
	id, err := suite.store.Create(suite.ctx, 42, false)
	suite.Require().NoError(err)

	suite.Require().NoError(suite.store.Delete(suite.ctx, id))

	_, err = suite.store.Validate(suite.ctx, id)
	suite.ErrorIs(err, sessions.ErrNotFound)
}

func (suite *StoreTestSuite) TestDeleteUnknownSessionIsNoOp() {
	suite.NoError(suite.store.Delete(suite.ctx, "does-not-exist"))
}
