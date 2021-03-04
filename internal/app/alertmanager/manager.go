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
	EventStreamer    smee.EventStreamer
	DeviceStateStore smee.DeviceStateStore
	AlertConfigs     map[string]smee.AlertConfig
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

	return group.Wait()
}

func (m *Manager) create(ctx context.Context, alert smee.Alert) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := m.AlertStore.CreateAlert(ctx, alert); err != nil {
		// TODO log err
	}
}

func (m *Manager) update(ctx context.Context, alert smee.Alert) {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := m.AlertStore.UpdateAlert(ctx, alert); err != nil {
		// TODO log err
	}
}
