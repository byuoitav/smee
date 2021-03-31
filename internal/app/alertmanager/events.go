package alertmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
)

func (m *Manager) generateEventAlerts(ctx context.Context) error {
	// stream setup timeout?
	stream, err := m.EventStreamer.Stream(ctx)
	if err != nil {
		return fmt.Errorf("unable to start event stream: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			event, err := stream.Next(ctx)
			if err != nil {
				return fmt.Errorf("unable to get next event: %w", err)
			}

			for typ, config := range m.AlertConfigs {
				trans := config.Create.Event
				if trans == nil {
					continue
				}

				if trans.Key != nil && !trans.Key.MatchString(event.Key) {
					continue
				}

				if trans.Value != nil && !trans.Value.MatchString(event.Value) {
					continue
				}

				alert := smee.Alert{
					Room:   event.Room,
					Device: event.Device,
					Type:   typ,
					Start:  time.Now(),
				}

				m.queue <- alertAction{
					action: "create",
					alert:  alert,
					events: []smee.IssueEvent{
						{
							Type:      "system-message",
							Timestamp: time.Now(),
							Data:      []byte(fmt.Sprintf(`{"msg": "|%v| %v alert started. Value: %v"}`, event.Device, typ, event.Value)),
						},
					},
				}
			}
		}
	}
}

func (m *Manager) closeEventAlerts(ctx context.Context) error {
	// stream setup timeout?
	stream, err := m.EventStreamer.Stream(ctx)
	if err != nil {
		return fmt.Errorf("unable to start event stream: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			event, err := stream.Next(ctx)
			if err != nil {
				return fmt.Errorf("unable to get next event: %w", err)
			}

			alerts, err := m.AlertStore.ActiveAlerts(ctx)
			if err != nil {
				// TODO log instead of fail?
				return fmt.Errorf("unable to get active alerts: %w", err)
			}

			for i := range alerts {
				alert := alerts[i]
				config, ok := m.AlertConfigs[alert.Type]
				if !ok {
					// TODO log that i don't know how to handle this alert
					continue
				}

				// make sure it's an event alert
				trans := config.Close.Event
				if trans == nil {
					continue
				}

				if event.Device != alert.Device {
					continue
				}

				if trans.Key != nil && !trans.Key.MatchString(event.Key) {
					continue
				}

				if trans.Value != nil && !trans.Value.MatchString(event.Value) {
					continue
				}

				// close the alert
				m.queue <- alertAction{
					action: "close",
					alert:  alert,
					events: []smee.IssueEvent{
						{
							Type:      "system-message",
							Timestamp: time.Now(),
							Data:      []byte(fmt.Sprintf(`{"msg": "|%v| %v alert ended. Value: %v"}`, event.Device, alert.Type, event.Value)),
						},
					},
				}
			}
		}
	}
}
