package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/matryer/is"
)

func TestGetActiveIssue(t *testing.T) {
	is := is.New(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	c, err := New(ctx, "")
	is.NoErr(err)

	iss, err := c.ActiveIssue(ctx, "ITB-1010")
	is.NoErr(err)
	fmt.Printf("issue: %+v\n", iss)
	is.Equal(iss, smee.Issue{})
}
