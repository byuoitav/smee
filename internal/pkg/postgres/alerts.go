package postgres

import (
	"context"

	"github.com/byuoitav/smee/internal/smee"
)

func (c *Client) CreateAlert(ctx context.Context, alert smee.Alert) (smee.Issue, error) {
	// get the current issue for the room
	// if exists, create alert for that issue
	// if not, create new issue and alert
	return smee.Issue{}, nil
}

func (c *Client) CloseAlert(ctx context.Context, issueID, alertID string) (smee.Issue, error) {
	return smee.Issue{}, nil
}

func (c *Client) LinkIncident(ctx context.Context, issueID string, inc smee.Incident) (smee.Issue, error) {
	return smee.Issue{}, nil
}

func (c *Client) AddIssueEvents(ctx context.Context, issueID string, events ...smee.IssueEvent) error {
	return nil
}
