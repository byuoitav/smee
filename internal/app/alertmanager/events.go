package alertmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"go.uber.org/zap"
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
				switch {
				case trans == nil:
					continue
				case trans.KeyMatches != nil && !trans.KeyMatches.MatchString(event.Key):
					continue
				case trans.KeyDoesNotMatch != nil && trans.KeyDoesNotMatch.MatchString(event.Key):
					continue
				case trans.ValueMatches != nil && !trans.ValueMatches.MatchString(event.Value):
					continue
				case trans.ValueDoesNotMatch != nil && trans.ValueDoesNotMatch.MatchString(event.Value):
					continue
				}

				alert := smee.Alert{
					Device: smee.Device{
						ID: event.DeviceID,
						Room: smee.Room{
							ID: event.RoomID,
						},
					},
					Type:  typ,
					Start: time.Now(),
				}

				m.queue <- alertAction{
					action: "create",
					alert:  alert,
					events: []smee.IssueEvent{
						{
							Type:      smee.TypeSystemMessage,
							Timestamp: time.Now(),
							Data:      smee.NewSystemMessage(fmt.Sprintf("AV Bot: |%v| %v alert started (Value: %v)", event.DeviceID, typ, event.Value)),
						},
					},
				}
			}
		}
	}
}

func (m *Manager) closeEventAlerts(ctx context.Context) error {
	// TODO stream setup timeout?
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

			alerts, err := m.IssueStore.ActiveAlerts(ctx)
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

				trans := config.Close.Event
				switch {
				case trans == nil:
					continue
				case event.DeviceID != alert.Device.ID || event.RoomID != alert.Device.Room.ID:
					continue
				case trans.KeyMatches != nil && !trans.KeyMatches.MatchString(event.Key):
					continue
				case trans.KeyDoesNotMatch != nil && trans.KeyDoesNotMatch.MatchString(event.Key):
					continue
				case trans.ValueMatches != nil && !trans.ValueMatches.MatchString(event.Value):
					continue
				case trans.ValueDoesNotMatch != nil && trans.ValueDoesNotMatch.MatchString(event.Value):
					continue
				}

				m.Log.Debug("Closing issue because of event", zap.Any("event", event))

				// close the alert
				m.queue <- alertAction{
					action: "close",
					alert:  alert,
					events: []smee.IssueEvent{
						{
							Type:      smee.TypeSystemMessage,
							Timestamp: time.Now(),
							Data:      smee.NewSystemMessage(fmt.Sprintf("AV Bot: |%v| %v alert ended (Value: %v)", event.DeviceID, alert.Type, event.Value)),
						},
					},
				}
			}
		}
	}
}
