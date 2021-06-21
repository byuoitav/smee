package smee

import (
	"context"
	"regexp"
	"time"
)

type IncidentStore interface {
	Incident(context.Context, string) (Incident, error)
	IncidentByName(context.Context, string) (Incident, error)
	AddIssueEvents(context.Context, string, ...IssueEvent) error
	CreateIncident(context.Context, Incident) (Incident, error)
}

type Incident struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// these fields are not filled in the issue store
	// only when you use the incident store
	Caller           string `json:"caller,omitempty"`
	ShortDescription string `json:"shortDescription,omitempty"`
}

type Event struct {
	RoomID   string
	DeviceID string
	Key      string
	Value    string
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
	// RunAlertQueries runs all store-specific queries and returns a map of queryName -> device id's that match the query
	RunAlertQueries(ctx context.Context) (map[string][]Device, error)
}

type Room struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Device struct {
	ID   string `json:"id"`
	Name string `json:"name"`

	// Room is the room a device is in
	Room Room `json:"room"`
}

type AlertConfig struct {
	// TODO Only have a a close for event alerts
	Create AlertTransition
	Close  AlertTransition
}

type AlertTransition struct {
	Event *AlertTransitionEvent
}

type AlertTransitionEvent struct {
	KeyMatches        *regexp.Regexp
	KeyDoesNotMatch   *regexp.Regexp
	ValueMatches      *regexp.Regexp
	ValueDoesNotMatch *regexp.Regexp
}

// change to room/device ID's
type Alert struct {
	ID      string    `json:"id"`
	IssueID string    `json:"issueID"`
	Device  Device    `json:"device"`
	Type    string    `json:"type"`
	Start   time.Time `json:"start"`
	End     time.Time `json:"end"`
}

func (a *Alert) Active() bool {
	return a.End.IsZero()
}

type AlertManager interface {
	Run(context.Context) error
}
