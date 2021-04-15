package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type StateStore struct {
	rdb            *redis.Client
	queries        map[string]query
	queryBatchSize int
}

type query func(key string, val []byte) bool

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
		rdb:            rdb,
		queries:        defaultQueries(),
		queryBatchSize: 128,
	}, nil
}

func (s *StateStore) AllQueries(ctx context.Context) (map[string][]string, error) {
	start := time.Now()

	// res := make(map[string][]string)
	runQueries := func(keys []string) error {
		vals, err := s.rdb.MGet(ctx, keys...).Result()
		if err != nil {
			return fmt.Errorf("unable to mget: %w", err)
		}

		for i, key := range keys {
			switch val := vals[i].(type) {
			case string:
				// TODO run query
				// fmt.Printf("%s (%T): %+v\n", key, val, val)
			default:
				fmt.Printf("%s (%T): %+v\n", key, val, val)
			}
		}

		return nil
	}

	var keyBatch []string

	iter := s.rdb.Scan(ctx, 0, "", int64(s.queryBatchSize)).Iterator()
	for iter.Next(ctx) {
		keyBatch = append(keyBatch, iter.Val())

		if len(keyBatch) == s.queryBatchSize {
			if err := runQueries(keyBatch); err != nil {
				return nil, fmt.Errorf("unable to run queries: %w", err)
			}

			keyBatch = nil
		}
	}

	fmt.Printf("took: %s\n", time.Since(start))
	return nil, nil
}
