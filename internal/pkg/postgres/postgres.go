package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type Client struct {
	pool *pgxpool.Pool
}

type issue struct {
	ID          int
	CouchRoomID string
	StartTime   time.Time // TODO should these be *time.Time?
	EndTime     time.Time
}

type alert struct {
	ID            int
	IssueID       int
	CouchRoomID   string
	CouchDeviceID string
	AlertType     string
	StartTime     time.Time // TODO should these be *time.Time?
	EndTime       time.Time
}

type incidentMapping struct {
	IssueID        int
	snSysID        string
	snTicketNumber string
}

type issueEvent struct {
	ID        int
	IssueID   int
	Time      time.Time // TODO should this be *time.Time?
	EventType string
	Data      json.RawMessage // TODO double check this type
}

type maintenanceInfo struct {
	CouchRoomID string
	StartTime   time.Time
	EndTime     time.Time // TODO need to rerun migrations to fix type
}

func New(ctx context.Context, connString string) (*Client, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connString: %w", err)
	}

	config.MaxConns = 32 // totally random...

	pool, err := pgxpool.ConnectConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to connect: %w", err)
	}

	return &Client{
		pool: pool,
	}, nil
}

func (c *Client) Close() error {
	c.pool.Close()
	return nil
}
