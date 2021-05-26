package postgres

import (
	"context"
	"fmt"

	"github.com/byuoitav/smee/internal/smee"
)

func (c *Client) ActiveIssue(ctx context.Context, roomID string) (smee.Issue, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to start tx: %w", err)
	}
	defer tx.Rollback(ctx)

	issID, err := c.activeIssueID(ctx, tx, roomID)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get active issue ID: %w", err)
	}

	smeeIss, err := c.smeeIssue(ctx, tx, issID)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get smeeIssue: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return smee.Issue{}, fmt.Errorf("unable to commit tx: %w", err)
	}

	return smeeIss, nil
}

func (c *Client) ActiveIssues(ctx context.Context) ([]smee.Issue, error) {
	return []smee.Issue{}, nil
}

