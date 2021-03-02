package alertmanager

import (
	"context"

	"github.com/byuoitav/smee/internal/smee"
	"golang.org/x/sync/errgroup"
)

// Manager ...
// Need a process to create alerts and other to close them
type Manager struct {
	AlertStore    smee.AlertStore
	EventStreamer smee.EventStreamer
	AlertConfigs  map[string]smee.AlertConfig
}

func (m *Manager) Run(ctx context.Context) error {
	group, gctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return m.generateEventAlerts(gctx)
	})

	group.Go(func() error {
		return m.closeEventAlerts(gctx)
	})

	return group.Wait()
}
