package alertmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
)

func (m *Manager) generateEventAlerts(ctx context.Context) error {
	events, err := m.EventStreamer.Stream(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case event := <-events:
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
					Messages: []smee.AlertMessage{
						{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("|%v| Alert started on %v. Value: %v", typ, event.Device, event.Value),
						},
					},
				}

				// TODO should i create in another goroutine?
				if err := m.AlertStore.Create(ctx, alert); err != nil {
					// TODO log err
				}
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
			alerts, err := m.AlertStore.Open(ctx)
			if err != nil {
				// TODO log
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
				alert.End = time.Now()
				alert.Messages = append(alert.Messages, smee.AlertMessage{
					Timestamp: time.Now(),
					Message:   fmt.Sprintf("|%v| Alert ended on %v. Value: %v", alert.Type, alert.Device, event.Value),
				})

				// TODO do in another goroutine?
				if err := m.AlertStore.Update(ctx, alert); err != nil {
					// TODO handle error?
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
