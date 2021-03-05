package smee

import (
	"context"
	"encoding/json"
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
	// AddEvent(ctx context.Context, issueID, comment string) error
}

type Issue struct {
	ID             string
	Room           string
	Start          time.Time
	End            time.Time
	ActiveAlerts   []Alert
	InactiveAlerts []Alert
	Events         []IssueEvent
}

type IssueEvent struct {
	Timestamp time.Time
	Type      string
	Data      json.RawMessage
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
	// Stream streams until ctx is cancelled
	Stream(ctx context.Context) (<-chan Event, error)
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
