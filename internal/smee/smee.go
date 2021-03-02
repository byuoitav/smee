package smee

import (
	"context"
	"regexp"
	"time"
)

type AlertStore interface {
	OpenAlerts(context.Context) []Alert
	UpdateAlert(context.Context, Alert)
}

type Event struct {
	DeviceID string
	Key      string
	Value    string
}

// EventStreamer ...
// TODO Probably need some wrapper for this so only one connection is created
type EventStreamer interface {
	// Stream streams until ctx is cancelled
	Stream(ctx context.Context) (<-chan Event, error)
}

type IssueStore interface {
}

type DeviceStateCache interface {
}

type AlertConfig struct {
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

type Alert struct {
	ID       string
	IssueID  string
	DeviceID string
	Start    time.Time
	End      time.Time
	Type     string

	// TODO this should live in alertmanager's version of alert
	Close AlertTransition
}

type AlertManager interface {
	Run(context.Context) error
	Manage(Alert)
}
