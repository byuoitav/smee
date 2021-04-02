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
	ActiveAlert(ctx context.Context, room, device, typ string) (Alert, bool, error)
	ActiveAlertsByType(context.Context, string) ([]Alert, error)
}

type IssueStore interface {
	CreateIssue(context.Context, Issue) (Issue, error)
	CloseIssue(context.Context, string) error
	ActiveIssues(context.Context) ([]Issue, error)
	ActiveIssueForRoom(context.Context, string) (Issue, bool, error)
	AddIssueEvent(context.Context, string, IssueEvent) error
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
	KeyMatches        *regexp.Regexp
	KeyDoesNotMatch   *regexp.Regexp
	ValueMatches      *regexp.Regexp
	ValueDoesNotMatch *regexp.Regexp
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
	// Manage(Alert) what was this for?
}
