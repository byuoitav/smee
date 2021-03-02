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
	AlertStore    smee.AlertStore
	EventStreamer smee.EventStreamer
	AlertConfigs  map[string]smee.AlertConfig
}

func (m *Manager) Run(ctx context.Context) error {
	group, gctx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return m.generateEventAlerts(gctx)
	})

	return group.Wait()
}

func (m *Manager) generateEventAlerts(ctx context.Context) error {
	events, err := m.EventStreamer.Stream(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-events:
			for typ, config := range m.AlertConfigs {
				if config.Create.Event == nil {
					continue
				}

				trans := config.Create.Event
				if trans.Key != nil && !trans.Key.MatchString(event.Key) {
					continue
				}

				if trans.Value != nil && !trans.Value.MatchString(event.Value) {
					continue
				}

				alert := smee.Alert{
					DeviceID: event.DeviceID,
					Start:    time.Now(),
					Type:     typ,
				}

				switch {
				case config.Close.Event != nil:
					alert.Close.Event = config.Close.Event
				case config.Close.StateQuery != nil:
					alert.Close.StateQuery = config.Close.StateQuery
				default:
					// TODO don't create this alert? no way to close it?
					// actually, probably should create it, manual close
				}

				// TODO create the alert
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (m *Manager) closeEventAlerts(ctx context.Context) error {
	events, err := m.EventStreamer.Stream(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-events:
			alerts := m.AlertStore.OpenAlerts(ctx)

			for i := range alerts {
				alert := alerts[i]
				trans := alert.Close.Event
				if trans == nil {
					continue
				}

				if event.DeviceID != alert.DeviceID {
					continue
				}

				if trans.Key != nil && !trans.Key.MatchString(event.Key) {
					continue
				}

				if trans.Value != nil && !trans.Value.MatchString(event.Value) {
					continue
				}

				// close the alert
				alert.End = time.Now()
				m.AlertStore.UpdateAlert(ctx, alert)
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
