package redis

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
)

type StateStore struct {
	rdb *redis.Client
}

func New(ctx context.Context, redisURL string) (*StateStore, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("unable to parse redisURL: %w", err)
	}

	rdb := redis.NewClient(opt)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping: %w", err)
	}

	return &StateStore{
		rdb: rdb,
	}, nil
}

func (s *StateStore) Query(ctx context.Context, query string) ([]string, error) {
	return nil, nil
}
