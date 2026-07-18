package tokens

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go-web-template/internal/config"

	"github.com/redis/go-redis/v9"
)

// Purposes namespace tokens so a confirm token can never be consumed as a reset
// token (and vice versa).
const (
	PurposeConfirm = "confirm"
	PurposeReset   = "reset"
)

var ErrNotFound = errors.New("token not found")

type TokenStoreInterface interface {
	Create(ctx context.Context, purpose string, userID int64, ttl time.Duration) (string, error)
	Consume(ctx context.Context, purpose, id string) (int64, error)
}

type Store struct {
	rdb *redis.Client
	cfg config.TokenConfig
}

var _ TokenStoreInterface = (*Store)(nil)

func NewStore(rdb *redis.Client, cfg config.TokenConfig) *Store {
	return &Store{rdb: rdb, cfg: cfg}
}

func (s *Store) Create(ctx context.Context, purpose string, userID int64, ttl time.Duration) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate token id: %w", err)
	}
	id := base64.RawURLEncoding.EncodeToString(buf)

	pipe := s.rdb.TxPipeline()
	pipe.HSet(ctx, key(purpose, id), map[string]any{
		"user_id":    userID,
		"created_at": time.Now().Unix(),
	})
	pipe.Expire(ctx, key(purpose, id), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("failed to store token: %w", err)
	}

	return id, nil
}

// Consume deletes the token after reading it, so each token works exactly once.
func (s *Store) Consume(ctx context.Context, purpose, id string) (int64, error) {
	val, err := s.rdb.HGet(ctx, key(purpose, id), "user_id").Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read token: %w", err)
	}

	if err := s.rdb.Del(ctx, key(purpose, id)).Err(); err != nil {
		return 0, fmt.Errorf("failed to delete token: %w", err)
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse token user id: %w", err)
	}

	return userID, nil
}

func key(purpose, id string) string {
	return "token:" + purpose + ":" + id
}
