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

	// TODO should this return the updated issue? to update the cache?
	AddIssueEvents(ctx context.Context, issueID string, event ...IssueEvent) error
	LinkIncident(ctx context.Context, issueID string, inc Incident) (Issue, error)

	ActiveIssue(ctx context.Context, roomID string) (Issue, error)
	ActiveIssues(context.Context) ([]Issue, error)
	CloseAlertsForIssue(ctx context.Context, issueID string) (Issue, error)
	AcknowledgeIssue(ctx context.Context, issueID string) (Issue, error)
	SetIssueStatus(ctx context.Context, issueID string, status string) (Issue, error)
	UnacknowledgeIssue(ctx context.Context, issueID string) (Issue, error)

	ActiveAlertExists(ctx context.Context, roomID, deviceID, typ string) (bool, error)
	ActiveAlerts(context.Context) ([]Alert, error)
	ActiveAlertsByType(context.Context, string) ([]Alert, error)
}

type Issue struct {
	// ID is the unique ID of this Issue
	ID string `json:"id"`

	// Room is the room this issue is associated with
	Room Room `json:"room"`

	// Start is the time this issue was created
	Start time.Time `json:"start"`

	// End is the time this issue was resolved
	End time.Time `json:"end"`

	// Alerts is a map of alertID -> alert
	Alerts map[string]Alert `json:"alerts"` // why a map...? i forgot -danny

	// Incidents is a map of incidentID -> incident
	Incidents map[string]Incident `json:"incidents"` // why a map...? i forgot -danny

	// Events is an ordered list by time of IssueEvents that have happened
	// on this Issue
	Events []IssueEvent `json:"events"`

	// Who Acknowledged the issue
	Acknowledged_By string `json:"acknowledged_by"`

	// Time the issue was acknowledged
	Acknowledged_Time time.Time `json:"acknowledged_time"`

	// Issue status
	Status string `json:"status"`
}

// Active returns true if this issue is currently active, and false if this
// has been closed.
func (i *Issue) Active() bool {
	return i.End.IsZero()
}

func (i *Issue) Acknowledged() bool {
	for _, element := range i.Alerts {
		if element.Acknowledged_Time.IsZero() {
			return true
		}
	}
	return false
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
