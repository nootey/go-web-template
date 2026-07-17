package sessions

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

const CookieName = "session"

var ErrNotFound = errors.New("session not found")

type Store struct {
	rdb *redis.Client
	cfg config.SessionConfig
}

func NewStore(rdb *redis.Client, cfg config.SessionConfig) *Store {
	return &Store{rdb: rdb, cfg: cfg}
}

// TTL is the session lifetime, and doubles as the cookie Max-Age. There is no
// sliding expiration: a slide could not outlive the cookie set at login.
func (s *Store) TTL(rememberMe bool) time.Duration {
	if rememberMe {
		return time.Duration(s.cfg.RememberMeTTLHours) * time.Hour
	}
	return time.Duration(s.cfg.TTLHours) * time.Hour
}

func (s *Store) Create(ctx context.Context, userID int64, rememberMe bool) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate session id: %w", err)
	}
	id := base64.RawURLEncoding.EncodeToString(buf)

	ttl := s.TTL(rememberMe)
	pipe := s.rdb.TxPipeline()
	pipe.HSet(ctx, key(id), map[string]any{
		"user_id":    userID,
		"created_at": time.Now().Unix(),
	})
	pipe.Expire(ctx, key(id), ttl)
	if _, err := pipe.Exec(ctx); err != nil {
		return "", fmt.Errorf("failed to store session: %w", err)
	}

	return id, nil
}

// Validate returns the session's user ID, or ErrNotFound if the session is
// missing or has expired.
func (s *Store) Validate(ctx context.Context, id string) (int64, error) {
	val, err := s.rdb.HGet(ctx, key(id), "user_id").Result()
	if errors.Is(err, redis.Nil) {
		return 0, ErrNotFound
	}
	if err != nil {
		return 0, fmt.Errorf("failed to read session: %w", err)
	}

	userID, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse session user id: %w", err)
	}

	return userID, nil
}

func (s *Store) Delete(ctx context.Context, id string) error {
	if err := s.rdb.Del(ctx, key(id)).Err(); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}

func key(id string) string {
	return "session:" + id
}
