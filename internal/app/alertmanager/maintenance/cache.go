package maintenance

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"go.uber.org/zap"
)

type Cache struct {
	MaintenanceStore smee.MaintenanceStore
	Log              *zap.Logger

	rooms   map[string]smee.MaintenanceInfo
	roomsMu sync.RWMutex
}

func (c *Cache) Sync(ctx context.Context) error {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()

	c.rooms = make(map[string]smee.MaintenanceInfo)

	if c.MaintenanceStore != nil {
		rooms, err := c.MaintenanceStore.RoomsInMaintenance(ctx)
		if err != nil {
			return fmt.Errorf("unable to get rooms in maintenance: %w", err)
		}

		for k, v := range rooms {
			c.rooms[k] = v
		}
	}

	// TODO remove
	c.rooms["ITB-1010"] = smee.MaintenanceInfo{
		RoomID: "ITB-1010",
		Start:  time.Now(),
		End:    time.Now().Add(1 * time.Hour),
	}

	c.Log.Info("Synced cache", zap.Int("roomsInMaintenance", len(c.rooms)))
	return nil
}

func (c *Cache) RoomsInMaintenance(ctx context.Context) (map[string]smee.MaintenanceInfo, error) {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()

	rooms := make(map[string]smee.MaintenanceInfo)
	for k, v := range c.rooms {
		if v.Enabled() {
			rooms[k] = v
		}
	}

	return rooms, nil
}

func (c *Cache) RoomMaintenanceInfo(ctx context.Context, room string) (smee.MaintenanceInfo, error) {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()
	return c.rooms[room], nil
}

func (c *Cache) SetMaintenanceInfo(ctx context.Context, info smee.MaintenanceInfo) error {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()

	if c.MaintenanceStore != nil {
		if err := c.MaintenanceStore.SetMaintenanceInfo(ctx, info); err != nil {
			return fmt.Errorf("unable to set maintenance info on substore: %w", err)
		}

		c.rooms[info.RoomID] = info
		return nil
	}

	c.rooms[info.RoomID] = info
	return nil
}
