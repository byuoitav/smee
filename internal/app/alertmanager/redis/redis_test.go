package redis

import (
	"context"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestAllQueries(t *testing.T) {
	is := is.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ss, err := New(ctx, "")
	is.NoErr(err)

	_, err = ss.AllQueries(ctx)
	is.NoErr(err)
}
