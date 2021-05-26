package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/matryer/is"
)

func TestCreateIssue(t *testing.T) {
	is := is.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := New(ctx, "")
	is.NoErr(err)

	tx, err := c.pool.Begin(ctx)
	is.NoErr(err)

	defer tx.Rollback(ctx)

	iss, err := c.createIssue(ctx, tx, issue{
		CouchRoomID: "ITB-DANNY",
		StartTime:   time.Now(),
	})
	is.NoErr(err)

	fmt.Printf("created: %+v\n", iss)

	// is.NoErr(tx.Commit(ctx))
}
