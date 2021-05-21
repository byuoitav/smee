package alertmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
)

func (m *Manager) manageStateAlerts(ctx context.Context) error {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			// figure out which devices should be alerting
			res, err := m.DeviceStateStore.RunAlertQueries(ctx)
			if err != nil {
				// TODO log
				continue
			}

			for typ, devices := range res {
				// get current open alerts for this query
				alerts, err := m.IssueStore.ActiveAlertsByType(ctx, typ)
				if err != nil {
					// TODO log
					continue
				}

				// build a map of the devices that _should_ have an alert
				shouldAlert := make(map[smee.DeviceInfo]bool, len(devices))
				for i := range devices {
					shouldAlert[devices[i]] = true
				}

				// build a map of the devices that currently have an alert
				curAlerts := make(map[smee.DeviceInfo]smee.Alert, len(alerts))
				for i := range alerts {
					curAlerts[smee.DeviceInfo{
						RoomID:   alerts[i].Room,
						DeviceID: alerts[i].Device,
					}] = alerts[i]
				}

				// create alerts for every device that should be alerting
				for dev := range shouldAlert {
					if _, ok := curAlerts[dev]; ok {
						// don't need to create an alert if it already exists
						continue
					}

					// create the alert
					alert := smee.Alert{
						Room:   dev.RoomID,
						Device: dev.DeviceID,
						Type:   typ,
						Start:  time.Now(),
					}

					m.queue <- alertAction{
						action: "create",
						alert:  alert,
						events: []smee.IssueEvent{
							{
								Type:      smee.TypeSystemMessage,
								Timestamp: time.Now(),
								Data:      smee.NewSystemMessage(fmt.Sprintf("AV Bot: |%v| %v alert started", dev.DeviceID, typ)),
							},
						},
					}
				}

				// close alerts for every device that are no longer alerting
				for device, alert := range curAlerts {
					if shouldAlert[device] {
						continue
					}

					// close the alert
					m.queue <- alertAction{
						action: "close",
						alert:  alert,
						events: []smee.IssueEvent{
							{
								Type:      smee.TypeSystemMessage,
								Timestamp: time.Now(),
								Data:      smee.NewSystemMessage(fmt.Sprintf("AV Bot: |%v| %v alert ended", device, typ)),
							},
						},
					}
				}
			}
		}
	}
}
