package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

type issueEvent struct {
	ID        int
	IssueID   int
	Time      time.Time
	EventType string
	Data      json.RawMessage
}

func (c *Client) issueEvents(ctx context.Context, tx pgx.Tx, issueID int) ([]issueEvent, error) {
	var events []issueEvent
	var event issueEvent

	_, err := tx.QueryFunc(ctx,
		"SELECT * FROM issue_events WHERE issue_id = $1",
		[]interface{}{issueID},
		[]interface{}{&event.ID, &event.IssueID, &event.Time, &event.EventType, &event.Data},
		func(pgx.QueryFuncRow) error {
			events = append(events, issueEvent{
				ID:        event.ID,
				IssueID:   event.IssueID,
				Time:      event.Time,
				EventType: event.EventType,
				Data:      event.Data,
			})

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return events, nil
}
