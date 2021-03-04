package issuemanager

import (
	"context"

	"github.com/byuoitav/smee/internal/smee"
)

var _ smee.AlertStore = &Manager{}

type Manager struct {
	smee.AlertStore
	IssueStore smee.IssueStore
}

func (m *Manager) CreateAlert(ctx context.Context, alert smee.Alert) error {
	iss, ok, err := m.IssueStore.ActiveIssueForRoom(ctx, alert.Room)
	if err != nil {
		// TODO handle error
	}

	if ok {
		// TODO update the issue
		return m.AlertStore.CreateAlert(ctx, alert)
	}

	// TODO create the issue

	return m.AlertStore.CreateAlert(ctx, alert)
}

func (m *Manager) UpdateAlert(ctx context.Context, alert smee.Alert) error {
	// we only care about this if the alert is closing
	if alert.End.IsZero() {
		return m.AlertStore.UpdateAlert(ctx, alert)
	}

	iss, ok, err := m.IssueStore.ActiveIssueForRoom(ctx, alert.Room)
	if err != nil {
		// TODO handle error
	}

	if ok {
		// TODO update the issue
		return m.AlertStore.UpdateAlert(ctx, alert)
	}

	// TODO create the issue

	return m.AlertStore.UpdateAlert(ctx, alert)
}