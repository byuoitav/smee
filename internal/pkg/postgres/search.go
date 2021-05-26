package postgres

import (
	"context"

	"github.com/byuoitav/smee/internal/smee"
)

// TODO support transactions for active issue

func (c *Client) ActiveIssue(ctx context.Context, roomID string) (smee.Issue, error) {
	if err := c.pool.QueryRow(ctx, "").Scan(); err != nil {
	}

	return smee.Issue{}, nil
}

func (c *Client) ActiveIssues(ctx context.Context) ([]smee.Issue, error) {
	return []smee.Issue{}, nil
}
