package alertmanager

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"golang.org/x/sync/errgroup"
)

func (m *Manager) manageStateAlerts(ctx context.Context) error {
	// create a goroutine to manage each state alert
	group, gctx := errgroup.WithContext(ctx)

	for t, c := range m.AlertConfigs {
		if c.Create.StateQuery == nil {
			continue
		}

		// create copies of loop variables so
		// we don't get the wrong value in the closure below
		typ := t
		config := c

		group.Go(func() error {
			return m.manageStateAlert(gctx, typ, config)
		})
	}

	// TODO add info to error?
	return group.Wait()
}

func (m *Manager) manageStateAlert(ctx context.Context, typ string, config smee.AlertConfig) error {
	// TODO just use the create interval/query?
	ticker := time.NewTicker(config.Create.StateQuery.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			devices, err := m.DeviceStateStore.Query(ctx, config.Create.StateQuery.Query)
			if err != nil {
				// TODO log
				continue
			}

			alerts, err := m.IssueStore.ActiveAlertsByType(ctx, typ)
			if err != nil {
				// TODO log
				continue
			}

			alerting := make(map[string]smee.Alert, len(alerts))
			for i := range alerts {
				alerting[alerts[i].Device] = alerts[i]
			}

			// close/create alerts
			for _, device := range devices {
				if alert, ok := alerting[device]; ok {
					// close the alert
					m.queue <- alertAction{
						action: "close",
						alert:  alert,
						events: []smee.IssueEvent{
							{
								Type:      "system-message",
								Timestamp: time.Now(),
								Data:      []byte(fmt.Sprintf(`{"msg": "|%v| %v alert ended."}`, device, typ)),
							},
						},
					}
					continue
				}

				// create the alert
				alert := smee.Alert{
					// TODO get the room!!!
					// Room:   event.Room
					Device: device,
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
							Data:      []byte(fmt.Sprintf(`{"msg": "|%v| %v alert started."}`, device, typ)),
						},
					},
				}
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
