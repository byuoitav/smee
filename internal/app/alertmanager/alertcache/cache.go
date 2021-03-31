package alertcache

import (
	"context"
	"fmt"
	"sync"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/segmentio/ksuid"
	"go.uber.org/zap"
)

type cache struct {
	persistent smee.AlertStore
	log        *zap.Logger

	cache map[string]smee.Alert
	sync.RWMutex
}

func New(ctx context.Context, persistent smee.AlertStore, log *zap.Logger) (*cache, error) {
	c := &cache{
		persistent: persistent,
		cache:      make(map[string]smee.Alert),
		log:        log,
	}

	if persistent != nil {
		alerts, err := persistent.ActiveAlerts(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to get persistent active alerts: %w", err)
		}

		for i := range alerts {
			c.cache[alerts[i].ID] = alerts[i]
		}
	}

	return c, nil
}

func (c *cache) CreateAlert(ctx context.Context, alert smee.Alert) (smee.Alert, error) {
	c.log.Info("Creating alert", zap.String("room", alert.Room), zap.String("device", alert.Device), zap.String("type", alert.Type))
	c.Lock()
	defer c.Unlock()

	switch {
	case alert.ID != "":
	case c.persistent != nil:
		var err error
		alert, err = c.persistent.CreateAlert(ctx, alert)
		if err != nil {
			return alert, fmt.Errorf("unable to create persistent alert: %w", err)
		}
	default:
		alert.ID = ksuid.New().String()
	}

	c.cache[alert.ID] = alert
	return alert, nil
}

func (c *cache) CloseAlert(ctx context.Context, id string) error {
	c.Lock()
	defer c.Unlock()

	if c.persistent != nil {
		if err := c.persistent.CloseAlert(ctx, id); err != nil {
			return fmt.Errorf("unable to close persistent alert: %w", err)
		}
	}

	delete(c.cache, id)
	return nil
}

func (c *cache) ActiveAlerts(ctx context.Context) ([]smee.Alert, error) {
	var res []smee.Alert

	c.RLock()
	defer c.RUnlock()

	for _, alert := range c.cache {
		res = append(res, alert)
	}

	return res, nil
}

func (c *cache) ActiveAlertsByType(ctx context.Context, typ string) ([]smee.Alert, error) {
	var res []smee.Alert

	c.RLock()
	defer c.RUnlock()

	for _, alert := range c.cache {
		if alert.Type == typ {
			res = append(res, alert)
		}
	}

	return res, nil
}
