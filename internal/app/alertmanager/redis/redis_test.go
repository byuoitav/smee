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

	queries, err := ss.RunQueries(ctx)
	is.NoErr(err)

	for query, ids := range queries {
		t.Logf("%s\n", query)
		for _, id := range ids {
			t.Logf("\t%s\n", id)
		}
	}
}
