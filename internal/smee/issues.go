package smee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type IssueStore interface {
	CreateAlert(context.Context, Alert) (Issue, error)
	CloseAlert(ctx context.Context, issueID, alertID string) (Issue, error)
	AddIssueEvents(ctx context.Context, issueID string, event ...IssueEvent) error
	LinkIncident(ctx context.Context, issueID string, inc Incident) (Issue, error)

	ActiveAlertExists(ctx context.Context, room, device, typ string) (bool, error)
	ActiveAlerts(context.Context) ([]Alert, error)
	ActiveAlertsByType(context.Context, string) ([]Alert, error)

	ActiveIssue(context.Context, string) (Issue, error)
	ActiveIssues(context.Context) ([]Issue, error)
}

type Issue struct {
	// ID is the unique ID of this Issue
	ID string `json:"id"`

	// Room is the room this issue is associated with
	Room string `json:"room"`

	// Start is the time this issue was created
	Start time.Time `json:"start"`

	// End is the time this issue was resolved
	End time.Time `json:"end"`

	// Alerts is a map of alertID -> alert
	Alerts map[string]Alert `json:"alerts"`

	// Incidents is a map of incidentID -> incident
	Incidents map[string]Incident `json:"incidents"`

	// Events is an ordered list by time of IssueEvents that have happened
	// on this Issue
	Events []IssueEvent `json:"events"`
}

// Active returns true if this issue is currently active, and false if this
// has been closed.
func (i *Issue) Active() bool {
	return i.End.IsZero()
}

type IssueEventType string

const (
	TypeSystemMessage IssueEventType = "system-message"
)

type SystemMessage struct {
	Message string `json:"msg"`
}

func NewSystemMessage(msg string) json.RawMessage {
	return []byte(fmt.Sprintf(`{"msg": "%s"}`, msg))
}

type IssueEvent struct {
	Timestamp time.Time       `json:"timestamp"`
	Type      IssueEventType  `json:"type"`
	Data      json.RawMessage `json:"data"`
}

func (i IssueEvent) ParseData() (interface{}, error) {
	switch i.Type {
	case TypeSystemMessage:
		var msg SystemMessage
		if err := json.Unmarshal(i.Data, &msg); err != nil {
			return nil, fmt.Errorf("unable to parse system message: %w", err)
		}

		return msg, nil
	default:
		return nil, errors.New("unknown type")
	}
}
