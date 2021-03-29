package smee

import (
	"context"
	"regexp"
	"time"
)

type AlertStore interface {
	CreateAlert(context.Context, Alert) (Alert, error)
	CloseAlert(context.Context, string) error
	ActiveAlerts(context.Context) ([]Alert, error)
	ActiveAlertsByType(context.Context, string) ([]Alert, error)
}

type IssueStore interface {
	CreateIssue(context.Context, Issue) (Issue, error)
	CloseIssue(context.Context, string) error
	ActiveIssueForRoom(context.Context, string) (Issue, bool, error)
	ActiveIssues(context.Context) (Issue, error)
}

type Issue struct {
	ID             string
	Room           string
	Start          time.Time
	End            time.Time
	ActiveAlerts   []Alert
	InactiveAlerts []Alert
}

type Event struct {
	Room   string
	Device string
	Key    string
	Value  string
}

// EventStreamer ...
// TODO Probably need some wrapper for this so only one connection is created
type EventStreamer interface {
	Stream(ctx context.Context) (EventStream, error)
}

type EventStream interface {
	Next(ctx context.Context) (Event, error)
	Close() error
}

type DeviceStateStore interface {
	// Query runs store-specific query and returns a list of DeviceID's that match the query
	Query(ctx context.Context, query string) ([]string, error)
}

type AlertConfig struct {
	// TODO Only have a a close for event alerts
	Create AlertTransition
	Close  AlertTransition
}

type AlertTransition struct {
	Event      *AlertTransitionEvent
	StateQuery *AlertTransitionStateQuery
}

type AlertTransitionEvent struct {
	Key   *regexp.Regexp
	Value *regexp.Regexp
}

type AlertTransitionStateQuery struct {
	Interval time.Duration
	Query    string
}

// change to room/device ID's
type Alert struct {
	ID      string
	IssueID string
	Room    string
	Device  string
	Type    string
	Start   time.Time
	End     time.Time
}

type AlertManager interface {
	Run(context.Context) error
	Manage(Alert)
}
