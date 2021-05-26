package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/jackc/pgx/v4"
)

type issue struct {
	ID          int
	CouchRoomID string
	StartTime   time.Time
	EndTime     *time.Time
}

func (c *Client) activeIssueID(ctx context.Context, tx pgx.Tx, roomID string) (int, error) {
	var id int

	err := tx.QueryRow(ctx,
		"SELECT id FROM issues WHERE couch_room_id = $1 AND end_time IS NULL LIMIT 1",
		roomID).Scan(&id)
	switch {
	case err == pgx.ErrNoRows:
		return 0, smee.ErrRoomIssueNotFound
	case err != nil:
		return 0, fmt.Errorf("unable to query/scan: %w", err)
	}

	return id, nil
}

func (c *Client) createIssue(ctx context.Context, tx pgx.Tx, iss issue) (issue, error) {
	err := tx.QueryRow(ctx,
		"INSERT INTO issues (couch_room_id, start_time) VALUES ($1, $2) RETURNING id",
		iss.CouchRoomID, iss.StartTime).Scan(&iss.ID)
	if err != nil {
		return issue{}, fmt.Errorf("unable to query/scan: %w", err)
	}

	return iss, nil
}

func (c *Client) issue(ctx context.Context, tx pgx.Tx, id int) (issue, error) {
	var iss issue

	err := tx.QueryRow(ctx,
		"SELECT * FROM issues WHERE id = $1",
		id).Scan(&iss.ID, &iss.CouchRoomID, &iss.StartTime, &iss.EndTime)
	if err != nil {
		return issue{}, fmt.Errorf("unable to get query/scan: %w", err)
	}

	return iss, nil
}

func (c *Client) smeeIssue(ctx context.Context, tx pgx.Tx, id int) (smee.Issue, error) {
	iss, err := c.issue(ctx, tx, id)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get issue: %w", err)
	}

	alerts, err := c.alerts(ctx, tx, id)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get alerts: %w", err)
	}

	incs, err := c.incidentMappings(ctx, tx, id)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get incidents: %w", err)
	}

	events, err := c.issueEvents(ctx, tx, id)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get events: %w", err)
	}

	smeeIss, err := buildIssue(iss, alerts, incs, events)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to build issue: %w", err)
	}

	return smeeIss, nil
}

func buildIssue(iss issue, alerts []alert, incs []incidentMapping, events []issueEvent) (smee.Issue, error) {
	smeeIss := smee.Issue{
		ID: strconv.Itoa(iss.ID),
		Room: smee.Room{
			ID:   iss.CouchRoomID,
			Name: iss.CouchRoomID,
		},
		Start:     iss.StartTime,
		End:       derefTime(iss.EndTime),
		Alerts:    make(map[string]smee.Alert),
		Incidents: make(map[string]smee.Incident),
	}

	for _, a := range alerts {
		smeeAlert := smee.Alert{
			ID:      strconv.Itoa(a.ID),
			IssueID: strconv.Itoa(a.IssueID),
			Device: smee.Device{
				ID:   a.CouchDeviceID,
				Name: a.CouchDeviceID,
				Room: smee.Room{
					ID:   a.CouchRoomID,
					Name: a.CouchRoomID,
				},
			},
			Type:  a.AlertType,
			Start: a.StartTime,
			End:   derefTime(a.EndTime),
		}

		smeeIss.Alerts[smeeAlert.ID] = smeeAlert
	}

	for _, inc := range incs {
		smeeInc := smee.Incident{
			ID:   inc.SNSysID,
			Name: inc.SNTicketNumber,
		}

		smeeIss.Incidents[smeeInc.ID] = smeeInc
	}

	for _, event := range events {
		smeeEvent := smee.IssueEvent{
			Timestamp: event.Time,
			Type:      smee.IssueEventType(event.EventType),
			Data:      event.Data,
		}

		smeeIss.Events = append(smeeIss.Events, smeeEvent)
	}

	return smeeIss, nil
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}

	return *t
}
