package smee

import (
	"context"
	"regexp"
	"time"
)

type AlertStore interface {
	Create(context.Context, Alert) error
	Update(context.Context, Alert) error
	Open(context.Context) ([]Alert, error)
	OpenByType(context.Context, string) ([]Alert, error)
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
	// Query runs store-specific query and returns a list of
	// DeviceID's that match the query
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

type Alert struct {
	ID       string
	Room     string
	Device   string
	Type     string
	Start    time.Time
	End      time.Time
	Messages []AlertMessage
}

type AlertMessage struct {
	Timestamp time.Time
	Message   string
}

type AlertManager interface {
	Run(context.Context) error
	Manage(Alert)
}
