package tokens_test

import (
	"context"
	"testing"
	"time"

	"go-web-template/internal/config"
	"go-web-template/internal/tokens"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type StoreTestSuite struct {
	suite.Suite
	mr    *miniredis.Miniredis
	store *tokens.Store
	ctx   context.Context
}

func TestStoreTestSuite(t *testing.T) {
	suite.Run(t, new(StoreTestSuite))
}

func (suite *StoreTestSuite) SetupTest() {
	suite.mr = miniredis.RunT(suite.T())
	rdb := redis.NewClient(&redis.Options{Addr: suite.mr.Addr()})
	suite.store = tokens.NewStore(rdb, config.TokenConfig{
		ConfirmTTLHours: 24,
		ResetTTLMinutes: 60,
	})
	suite.ctx = context.Background()
}

func (suite *StoreTestSuite) TestCreateConsumeRoundTrip() {
	id, err := suite.store.Create(suite.ctx, tokens.PurposeConfirm, 42, time.Hour)
	suite.Require().NoError(err)
	suite.NotEmpty(id)

	userID, err := suite.store.Consume(suite.ctx, tokens.PurposeConfirm, id)
	suite.Require().NoError(err)
	suite.Equal(int64(42), userID)
}

func (suite *StoreTestSuite) TestCreateGeneratesDistinctIDs() {
	first, err := suite.store.Create(suite.ctx, tokens.PurposeReset, 1, time.Hour)
	suite.Require().NoError(err)
	second, err := suite.store.Create(suite.ctx, tokens.PurposeReset, 1, time.Hour)
	suite.Require().NoError(err)

	suite.NotEqual(first, second)
}

func (suite *StoreTestSuite) TestCreateSetsTTL() {
	id, err := suite.store.Create(suite.ctx, tokens.PurposeConfirm, 7, 2*time.Hour)
	suite.Require().NoError(err)

	suite.Equal(2*time.Hour, suite.mr.TTL("token:"+tokens.PurposeConfirm+":"+id))
}

func (suite *StoreTestSuite) TestConsumeIsSingleUse() {
	id, err := suite.store.Create(suite.ctx, tokens.PurposeConfirm, 42, time.Hour)
	suite.Require().NoError(err)

	_, err = suite.store.Consume(suite.ctx, tokens.PurposeConfirm, id)
	suite.Require().NoError(err)

	_, err = suite.store.Consume(suite.ctx, tokens.PurposeConfirm, id)
	suite.ErrorIs(err, tokens.ErrNotFound)
}

func (suite *StoreTestSuite) TestConsumeWrongPurpose() {
	id, err := suite.store.Create(suite.ctx, tokens.PurposeConfirm, 42, time.Hour)
	suite.Require().NoError(err)

	_, err = suite.store.Consume(suite.ctx, tokens.PurposeReset, id)
	suite.ErrorIs(err, tokens.ErrNotFound)
}

func (suite *StoreTestSuite) TestConsumeUnknownToken() {
	_, err := suite.store.Consume(suite.ctx, tokens.PurposeConfirm, "does-not-exist")
	suite.ErrorIs(err, tokens.ErrNotFound)
}

func (suite *StoreTestSuite) TestConsumeExpiredToken() {
	id, err := suite.store.Create(suite.ctx, tokens.PurposeReset, 42, time.Hour)
	suite.Require().NoError(err)

	suite.mr.FastForward(time.Hour)

	_, err = suite.store.Consume(suite.ctx, tokens.PurposeReset, id)
	suite.ErrorIs(err, tokens.ErrNotFound)
}
