package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/jackc/pgx/v4/pgxpool"
	"go.uber.org/zap"
)

type Client struct {
	Log  *zap.Logger
	pool *pgxpool.Pool
}

type maintenanceInfo struct {
	CouchRoomID string
	StartTime   *time.Time
	EndTime     *time.Time
}

func New(ctx context.Context, connString string) (*Client, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connString: %w", err)
	}

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	return &Client{
		Log:  zap.NewNop(),
		pool: pool,
	}, nil
}

func (c *Client) Close() error {
	c.pool.Close()
	return nil
}

func (c *Client) CreateAlert(ctx context.Context, smeeAlert smee.Alert) (smee.Issue, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to start transaction: %w", err)
	}
	defer tx.Rollback(ctx) // what does rollback do if ctx is past deadline?

	issID, err := c.activeIssueID(ctx, tx, smeeAlert.Device.Room.ID)
	switch {
	case errors.Is(err, smee.ErrRoomIssueNotFound):
		// create a new issue
		iss := issue{
			CouchRoomID: smeeAlert.Device.Room.ID,
			StartTime:   smeeAlert.Start,
		}

		iss, err := c.createIssue(ctx, tx, iss)
		if err != nil {
			return smee.Issue{}, fmt.Errorf("unable to create issue: %w", err)
		}

		c.Log.Info("Created issue", zap.String("roomID", iss.CouchRoomID), zap.Int("issueID", issID))
	case err != nil:
		return smee.Issue{}, fmt.Errorf("unable to get active issue: %w", err)
	}

	// create the alert
	a := alert{
		IssueID:       issID,
		CouchRoomID:   smeeAlert.Device.Room.ID,
		CouchDeviceID: smeeAlert.Device.ID,
		AlertType:     smeeAlert.Type,
		StartTime:     smeeAlert.Start,
	}

	a, err = c.createAlert(ctx, tx, a)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to create alert: %w", err)
	}

	c.Log.Info("Created alert", zap.String("roomID", a.CouchRoomID), zap.Int("issueID", issID), zap.Int("alertID", a.ID), zap.String("deviceID", a.CouchDeviceID), zap.String("type", a.AlertType))

	// get the issue after creating this alert

	if err := tx.Commit(ctx); err != nil {
		return smee.Issue{}, fmt.Errorf("unable to commit transaction: %w", err)
	}

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
