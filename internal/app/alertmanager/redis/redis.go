package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type StateStore struct {
	rdb     *redis.Client
	queries map[string]query
}

type query func() bool

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
		rdb:     rdb,
		queries: defaultQueries(),
	}, nil
}

func (s *StateStore) Query(ctx context.Context, query string) ([]string, error) {
	return nil, nil
}

func (s *StateStore) AllQueries(ctx context.Context) (map[string][]string, error) {
	/*
		values := func(keys []string) error {
			for _, key := range keys {
				_, err := s.rdb.Get(ctx, key).Bytes()
				if err != nil {
					return fmt.Errorf("unable to get key %q: %w", key, err)
				}
			}

			return nil
		}
	*/

	mGetValues := func(keys []string) error {
		vals, err := s.rdb.MGet(ctx, keys...).Result()
		if err != nil {
			return fmt.Errorf("unable to mget: %w", err)
		}

		fmt.Printf("\tmGetValues %v: len(vals): %v\n", len(keys), len(vals))
		return nil
	}

	scan := func(size int64) error {
		start := time.Now()
		keys := []string{}
		count := 0

		iter := s.rdb.Scan(ctx, 0, "", size).Iterator()
		for iter.Next(ctx) {
			count++
			keys = append(keys, iter.Val())

			if int64(len(keys)) == size {
				if err := mGetValues(keys); err != nil {
					return fmt.Errorf("unable to get values: %w", err)
				}

				keys = nil
			}
		}
		if err := iter.Err(); err != nil {
			return fmt.Errorf("unable to iterate through keys: %w", err)
		}

		// get the last batch
		if err := mGetValues(keys); err != nil {
			return fmt.Errorf("unable to get values: %w", err)
		}

		fmt.Printf("scan %v: got %v keys in %s\n", size, count, time.Since(start))
		return nil
	}

	keys := func() error {
		// keys
		start := time.Now()

		keys, err := s.rdb.Keys(ctx, "*").Result()
		if err != nil {
			return fmt.Errorf("unable to get keys: %w", err)
		}

		if err := mGetValues(keys); err != nil {
			return fmt.Errorf("unable to mGet values: %w", err)
		}

		fmt.Printf("keys: got %v keys in %s\n", len(keys), time.Since(start))
		return nil
	}

	if err := keys(); err != nil {
		return nil, fmt.Errorf("err keys: %w", err)
	}

	if err := scan(256); err != nil {
		return nil, fmt.Errorf("err scan 256: %w", err)
	}

	return nil, nil
}
