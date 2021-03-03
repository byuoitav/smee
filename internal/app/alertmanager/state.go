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

			alerts, err := m.AlertStore.OpenByType(ctx, typ)
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
					alert.End = time.Now()
					alert.Messages = append(alert.Messages, smee.AlertMessage{
						Timestamp: time.Now(),
						Message:   fmt.Sprintf("|%v| Alert ended on %v.", alert.Type, alert.Device),
					})

					// TODO do in another goroutine?
					if err := m.AlertStore.Update(ctx, alert); err != nil {
						// TODO handle error?
					}

					continue
				}

				// create the alert
				alert := smee.Alert{
					// Room:   event.Room
					Device: device,
					Type:   typ,
					Start:  time.Now(),
					Messages: []smee.AlertMessage{
						{
							Timestamp: time.Now(),
							Message:   fmt.Sprintf("|%v| Alert started on %v.", typ, device),
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
