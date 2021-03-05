package alertmanager

import (
	"context"
	"time"

	"github.com/byuoitav/smee/internal/smee"
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

	queue chan alertAction
}

type alertAction struct {
	action string
	alert  smee.Alert
}

func (m *Manager) Run(ctx context.Context) error {

	group, gctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return m.generateEventAlerts(gctx)
	})

	group.Go(func() error {
		return m.manageStateAlerts(gctx)
	})

	group.Go(func() error {
		return m.closeEventAlerts(gctx)
	})

	group.Go(func() error {
		return m.runAlertActions(gctx)
	})

	return group.Wait()
}

// runAlertActions ensures that actions generated by this manager
// are run in order of their placement in the queue. this makes handling
// issue creation/closure much simpler
func (m *Manager) runAlertActions(ctx context.Context) error {
	m.queue = make(chan alertAction, 1024)

	for {
		select {
		case action := <-m.queue:
			switch action.action {
			case "create":
				m.createAlert(ctx, action.alert)
			case "close":
				m.closeAlert(ctx, action.alert)
			default:
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *Manager) createAlert(ctx context.Context, alert smee.Alert) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// see if an issue already exists for this room
	issue, ok, err := m.IssueStore.ActiveIssueForRoom(ctx, alert.Room)
	switch {
	case err != nil:
	case !ok:
		// create a new issue
		issue = smee.Issue{
			Start: time.Now(),
			Room:  alert.Room,
		}

		issue, err = m.IssueStore.CreateIssue(ctx, issue)
		if err != nil {
		}
	}

	// tie this alert to the active issue for this room
	alert.IssueID = issue.ID

	// TODO events

	// create the alert
	_, err = m.AlertStore.CreateAlert(ctx, alert)
	if err != nil {
	}
}

func (m *Manager) closeAlert(ctx context.Context, alert smee.Alert) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// get the issue for this room
	issue, ok, err := m.IssueStore.ActiveIssueForRoom(ctx, alert.Room)
	switch {
	case err != nil:
	case ok:
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
