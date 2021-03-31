package alertmanager

import (
	"context"
	"errors"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// Manager ...
// Need a process to create alerts and other to close them
type Manager struct {
	AlertStore       smee.AlertStore
	IssueStore       smee.IssueStore
	EventStreamer    smee.EventStreamer
	DeviceStateStore smee.DeviceStateStore
	AlertConfigs     map[string]smee.AlertConfig
	Log              *zap.Logger

	queue chan alertAction
}

type alertAction struct {
	action string
	alert  smee.Alert
	events []smee.IssueEvent
}

func (m *Manager) Run(ctx context.Context) error {
	m.queue = make(chan alertAction, 1024)
	group, gctx := errgroup.WithContext(ctx)

	switch {
	case m.AlertStore == nil:
		return errors.New("alert store required")
	case m.IssueStore == nil:
		return errors.New("issue store required")
	}

	group.Go(func() error {
		return m.runAlertActions(gctx)
	})

	group.Go(func() error {
		return m.manageStateAlerts(gctx)
	})

	group.Go(func() error {
		return m.generateEventAlerts(gctx)
	})

	group.Go(func() error {
		return m.closeEventAlerts(gctx)
	})

	m.Log.Info("Alert manager running")
	return group.Wait()
}

// runAlertActions ensures that actions generated by this manager
// are run in order of their placement in the queue. this makes handling
// issue creation/closure much simpler
func (m *Manager) runAlertActions(ctx context.Context) error {
	for {
		select {
		case action := <-m.queue:
			switch action.action {
			case "create":
				m.createAlert(ctx, action.alert, action.events)
			case "close":
				m.closeAlert(ctx, action.alert, action.events)
			default:
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *Manager) createAlert(ctx context.Context, alert smee.Alert, events []smee.IssueEvent) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// see if an issue already exists for this room
	issue, ok, err := m.IssueStore.ActiveIssueForRoom(ctx, alert.Room)
	switch {
	case err != nil:
	case !ok:
		// create a new issue
		issue = smee.Issue{
			Start: alert.Start,
			Room:  alert.Room,
		}

		issue, err = m.IssueStore.CreateIssue(ctx, issue)
		if err != nil {
		}
	}

	// tie this alert to the active issue for this room
	alert.IssueID = issue.ID

	// create the alert
	_, err = m.AlertStore.CreateAlert(ctx, alert)
	if err != nil {
	}

	// add all of the events
	for _, event := range events {
		if err := m.IssueStore.AddIssueEvent(ctx, issue.ID, event); err != nil {
		}
	}
}

func (m *Manager) closeAlert(ctx context.Context, alert smee.Alert, events []smee.IssueEvent) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// get the issue for this room
	issue, ok, err := m.IssueStore.ActiveIssueForRoom(ctx, alert.Room)
	switch {
	case err != nil:
	case ok:
		// add all of the events
		for _, event := range events {
			if err := m.IssueStore.AddIssueEvent(ctx, issue.ID, event); err != nil {
			}
		}

		// close this issue if no more alerts are open
		// and if closing this one should close the issue
		// TODO does it not close if a SN incident is attached to it?
		if len(issue.ActiveAlerts) == 1 && issue.ActiveAlerts[0].ID == alert.ID {
			if err := m.IssueStore.CloseIssue(ctx, issue.ID); err != nil {
			}
		}
	}

	// close the alert
	if err := m.AlertStore.CloseAlert(ctx, alert.ID); err != nil {
	}
}
