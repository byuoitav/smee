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

func (c *Client) inactiveIssueID(ctx context.Context, tx pgx.Tx, roomID string) (int, error) {
	var id int

	err := tx.QueryRow(ctx,
		"SELECT id FROM issues WHERE couch_room_id = $1 AND end_time IS NOT NULL LIMIT 1",
		roomID).Scan(&id)
	switch {
	case err == pgx.ErrNoRows:
		return 0, smee.ErrRoomIssueNotFound
	case err != nil:
		return 0, fmt.Errorf("unable to query/scan: %w", err)
	}

	return id, nil
}

func (c *Client) inactiveIssueIDs(ctx context.Context, tx pgx.Tx) ([]int, error) {
	var ids []int
	var id int

	_, err := tx.QueryFunc(ctx,
		"SELECT id FROM issues WHERE end_time IS NOT NULL",
		[]interface{}{},
		[]interface{}{&id},
		func(pgx.QueryFuncRow) error {
			tmp := id
			ids = append(ids, tmp)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return ids, nil
}

func (c *Client) IssuesID(ctx context.Context, tx pgx.Tx) ([]int, error) {
	var ids []int
	var id int

	_, err := tx.QueryFunc(ctx,
		"SELECT id FROM issues",
		[]interface{}{},
		[]interface{}{&id},
		func(pgx.QueryFuncRow) error {
			tmp := id
			ids = append(ids, tmp)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return ids, nil
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

func (c *Client) activeIssueIDs(ctx context.Context, tx pgx.Tx) ([]int, error) {
	var ids []int
	var id int

	_, err := tx.QueryFunc(ctx,
		"SELECT id FROM issues WHERE end_time IS NULL",
		[]interface{}{},
		[]interface{}{&id},
		func(pgx.QueryFuncRow) error {
			tmp := id // don't know if i actually need this or not
			ids = append(ids, tmp)
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return ids, nil
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

func (c *Client) closeIssue(ctx context.Context, tx pgx.Tx, issueID int) error {
	res, err := tx.Exec(ctx,
		"UPDATE issues SET end_time = $1 WHERE id = $2",
		time.Now(), issueID)
	switch {
	case err != nil:
		return fmt.Errorf("unable to exec: %w", err)
	case res.RowsAffected() == 0:
		return fmt.Errorf("invalid issueID")
	}

	return nil
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
		Alerts:    make(map[string]smee.Alert),
		Incidents: make(map[string]smee.Incident),
	}

	if iss.EndTime != nil {
		smeeIss.End = *iss.EndTime
	}

	for _, a := range alerts {
		smeeAlert := convertAlert(a)
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
