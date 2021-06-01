package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4"
)

type alert struct {
	ID            int
	IssueID       int
	CouchRoomID   string
	CouchDeviceID string
	AlertType     string
	StartTime     time.Time
	EndTime       *time.Time
}

func (c *Client) createAlert(ctx context.Context, tx pgx.Tx, a alert) (alert, error) {
	err := tx.QueryRow(ctx,
		"INSERT INTO alerts (issue_id, couch_room_id, couch_device_id, alert_type, start_time) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		a.IssueID, a.CouchRoomID, a.CouchDeviceID, a.AlertType, a.StartTime).Scan(&a.ID)
	if err != nil {
		return alert{}, fmt.Errorf("unable to query/scan: %w", err)
	}

	return a, nil
}

func (c *Client) closeAlert(ctx context.Context, tx pgx.Tx, alertID int) error {
	res, err := tx.Exec(ctx,
		"UPDATE alerts SET end_time = $1 WHERE id = $2",
		time.Now(), alertID)
	switch {
	case err != nil:
		return fmt.Errorf("unable to exec: %w", err)
	case res.RowsAffected() == 0:
		return fmt.Errorf("invalid alertID")
	}

	return nil
}

func (c *Client) queryAlerts(ctx context.Context, tx pgx.Tx, query string, args ...interface{}) ([]alert, error) {
	var alerts []alert
	var a alert

	_, err := tx.QueryFunc(ctx, query, args,
		[]interface{}{&a.ID, &a.IssueID, &a.CouchRoomID, &a.CouchDeviceID, &a.AlertType, &a.StartTime, &a.EndTime},
		func(pgx.QueryFuncRow) error {
			alerts = append(alerts, alert{
				ID:            a.ID,
				IssueID:       a.IssueID,
				CouchRoomID:   a.CouchRoomID,
				CouchDeviceID: a.CouchDeviceID,
				AlertType:     a.AlertType,
				StartTime:     a.StartTime,
				EndTime:       a.EndTime,
			})

			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return alerts, nil
}

func (c *Client) alerts(ctx context.Context, tx pgx.Tx, issueID int) ([]alert, error) {
	return c.queryAlerts(ctx, tx, "SELECT * FROM alerts WHERE issue_id = $1", issueID)
}

func (c *Client) activeAlertExists(ctx context.Context, tx pgx.Tx, roomID, deviceID, typ string) (bool, error) {
	alerts, err := c.queryAlerts(ctx, tx,
		"SELECT * FROM alerts WHERE end_time IS NULL AND couch_room_id = $1 AND couch_device_id = $2 AND alert_type = $3",
		roomID, deviceID, typ)
	if err != nil {
		return false, fmt.Errorf("unable to query alerts: %w", err)
	}

	return len(alerts) > 0, nil
}

func (c *Client) activeAlerts(ctx context.Context, tx pgx.Tx) ([]alert, error) {
	return c.queryAlerts(ctx, tx, "SELECT * FROM alerts WHERE end_time IS NULL")
}

func (c *Client) activeAlertsByType(ctx context.Context, tx pgx.Tx, typ string) ([]alert, error) {
	return c.queryAlerts(ctx, tx, "SELECT * FROM alerts WHERE end_time IS NULL AND alert_type = $1", typ)
}

func (c *Client) activeAlertCount(ctx context.Context, tx pgx.Tx, issueID int) (int, error) {
	var ids []int
	var id int

	_, err := tx.QueryFunc(ctx,
		"SELECT id FROM alerts WHERE issue_id = $1 AND end_time is NULL",
		[]interface{}{issueID},
		[]interface{}{&id},
		func(pgx.QueryFuncRow) error {
			ids = append(ids, id)
			return nil
		},
	)
	if err != nil {
		return 0, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return len(ids), nil
}
