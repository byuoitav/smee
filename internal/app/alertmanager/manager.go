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
	IssueStore       smee.IssueStore
	MaintenanceStore smee.MaintenanceStore
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

	if m.IssueStore == nil {
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
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// see if this alert already exists
	exists, err := m.IssueStore.ActiveAlertExists(ctx, alert.Room, alert.Device, alert.Type)
	switch {
	case err != nil:
		m.Log.Error("unable to check if active alert exists", zap.Error(err), zap.String("room", alert.Room), zap.String("device", alert.Device), zap.String("type", alert.Type))
		return
	case exists:
		// don't need to do anything, this alert already exists
		// TODO maybe add the event to the issue?
		return
	}

	issue, err := m.IssueStore.CreateAlert(ctx, alert)
	if err != nil {
		m.Log.Error("unable to create alert", zap.Error(err), zap.String("room", alert.Room), zap.String("device", alert.Device), zap.String("type", alert.Type))
		return
	}

	if err := m.IssueStore.AddIssueEvents(ctx, issue.ID, events...); err != nil {
		m.Log.Error("unable to add issue events", zap.Error(err), zap.String("issueID", issue.ID), zap.String("room", issue.Room))
		return
	}
}

func (m *Manager) closeAlert(ctx context.Context, alert smee.Alert, events []smee.IssueEvent) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	issue, err := m.IssueStore.CloseAlert(ctx, alert.IssueID, alert.ID)
	if err != nil {
		m.Log.Error("unable to close alert", zap.Error(err), zap.String("issueID", alert.IssueID), zap.String("alertID", alert.ID))
		return
	}

	if err := m.IssueStore.AddIssueEvents(ctx, issue.ID, events...); err != nil {
		m.Log.Error("unable to add issue events", zap.Error(err), zap.String("issueID", alert.IssueID), zap.String("alertID", alert.ID))
		return
	}
}
