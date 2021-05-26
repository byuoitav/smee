package postgres

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
)

// TODO support transactions for active issue

func (c *Client) ActiveIssue(ctx context.Context, roomID string) (smee.Issue, error) {
	var iss issue
	err := c.pool.QueryRow(ctx,
		"SELECT * FROM issues WHERE couch_room_id = $1 AND end_time IS NULL LIMIT 1",
		roomID).Scan(&iss.ID, &iss.CouchRoomID, &iss.StartTime, &iss.EndTime)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get issue: %w", err)
	}

	var alerts []alert
	var a alert
	_, err = c.pool.QueryFunc(ctx,
		"SELECT * FROM alerts WHERE issue_id = $1",
		[]interface{}{iss.ID},
		[]interface{}{&a.ID, &a.IssueID, &a.CouchRoomID, &a.CouchDeviceID, &a.AlertType, &a.StartTime, &a.EndTime},
		func(pgx.QueryFuncRow) error {
			alerts = append(alerts, alert{
				ID:            a.ID,
				IssueID:       a.IssueID,
				CouchRoomID:   a.CouchRoomID,
				CouchDeviceID: a.CouchDeviceID,
				AlertType:     a.AlertType,
				StartTime:     a.StartTime,
				EndTime:       a.EndTime,
			})
			return nil
		},
	)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get alerts: %w", err)
	}

	var incs []incidentMapping
	var inc incidentMapping
	_, err = c.pool.QueryFunc(ctx,
		"SELECT * FROM sn_incident_mappings WHERE issue_id = $1",
		[]interface{}{iss.ID},
		[]interface{}{&inc.IssueID, &inc.SNSysID, &inc.SNTicketNumber},
		func(pgx.QueryFuncRow) error {
			incs = append(incs, incidentMapping{
				IssueID:        inc.IssueID,
				SNSysID:        inc.SNSysID,
				SNTicketNumber: inc.SNTicketNumber,
			})
			return nil
		},
	)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get incident mappings: %w", err)
	}

	var events []issueEvent
	var event issueEvent
	_, err = c.pool.QueryFunc(ctx,
		"SELECT * FROM issue_events WHERE issue_id = $1",
		[]interface{}{iss.ID},
		[]interface{}{&event.ID, &event.IssueID, &event.Time, &event.EventType, &event.Data},
		func(pgx.QueryFuncRow) error {
			events = append(events, issueEvent{
				ID:        event.ID,
				IssueID:   event.IssueID,
				Time:      event.Time,
				EventType: event.EventType,
				Data:      event.Data,
			})
			return nil
		},
	)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to get issue events: %w", err)
	}

	smeeIss, err := buildIssue(iss, alerts, incs, events)
	if err != nil {
		return smee.Issue{}, fmt.Errorf("unable to build issue: %w", err)
	}

	return smeeIss, nil
}

func (c *Client) ActiveIssues(ctx context.Context) ([]smee.Issue, error) {
	return []smee.Issue{}, nil
}

func (c *Client) getIssueData(ctx context.Context, tx pgxpool.Tx, iss issue) (smee.Issue, error) {
	return smee.Issue{}, nil
}

func buildIssue(iss issue, alerts []alert, incs []incidentMapping, events []issueEvent) (smee.Issue, error) {
	smeeIss := smee.Issue{
		ID: strconv.Itoa(iss.ID),
		Room: smee.Room{
			ID:   iss.CouchRoomID,
			Name: iss.CouchRoomID,
		},
		Start:     iss.StartTime,
		End:       derefTime(iss.EndTime),
		Alerts:    make(map[string]smee.Alert),
		Incidents: make(map[string]smee.Incident),
	}

	for _, a := range alerts {
		smeeAlert := smee.Alert{
			ID:      strconv.Itoa(a.ID),
			IssueID: strconv.Itoa(a.IssueID),
			Device: smee.Device{
				ID:   a.CouchDeviceID,
				Name: a.CouchDeviceID,
				Room: smee.Room{
					ID:   a.CouchRoomID,
					Name: a.CouchRoomID,
				},
			},
			Type:  a.AlertType,
			Start: a.StartTime,
			End:   derefTime(a.EndTime),
		}

		smeeIss.Alerts[smeeAlert.ID] = smeeAlert
	}

	for _, inc := range incs {
		smeeInc := smee.Incident{
			ID:   inc.SNSysID,
			Name: inc.SNTicketNumber,
		}

		smeeIss.Incidents[smeeInc.ID] = smeeInc
	}

	for _, event := range events {
		smeeEvent := smee.IssueEvent{
			Timestamp: event.Time,
			Type:      smee.IssueEventType(event.EventType),
			Data:      event.Data,
		}

		smeeIss.Events = append(smeeIss.Events, smeeEvent)
	}

	return smeeIss, nil
}

func derefTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}

	return *t
}
