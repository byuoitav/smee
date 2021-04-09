package smee

import (
	"context"
	"encoding/json"
	"regexp"
	"time"
)

type IssueStore interface {
	CreateAlert(context.Context, Alert) (Issue, error)
	CloseAlert(ctx context.Context, issueID, alertID string) (Issue, error)
	AddIssueEvents(ctx context.Context, issueID string, event ...IssueEvent) error

	ActiveAlertExists(ctx context.Context, room, device, typ string) (bool, error)
	ActiveAlerts(context.Context) ([]Alert, error)
	ActiveAlertsByType(context.Context, string) ([]Alert, error)
	ActiveIssues(context.Context) ([]Issue, error)
}

type Issue struct {
	ID    string
	Room  string
	Start time.Time
	End   time.Time

	// Alerts is a map of an alertID to an alert
	Alerts map[string]Alert
	Events []IssueEvent
}

func (i *Issue) Active() bool {
	return i.End.IsZero()
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

func (a *Alert) Active() bool {
	return a.End.IsZero()
}

type AlertManager interface {
	Run(context.Context) error
	// Manage(Alert) what was this for?
}
