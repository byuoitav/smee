package maintenance

import (
	"context"
	"fmt"
	"sync"

	"github.com/byuoitav/smee/internal/smee"
	"go.uber.org/zap"
)

type Cache struct {
	MaintenanceStore smee.MaintenanceStore
	Log              *zap.Logger

	rooms   map[string]bool
	roomsMu sync.RWMutex
}

func (c *Cache) Sync(ctx context.Context) error {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()

	c.rooms = make(map[string]bool)

	if c.MaintenanceStore != nil {
		rooms, err := c.RoomsInMaintenance(ctx)
		if err != nil {
			return fmt.Errorf("unable to get rooms in maintenance: %w", err)
		}

		for i := range rooms {
			c.rooms[rooms[i]] = true
		}
	}

	c.Log.Info("Synced cache", zap.Int("roomsInMaintenance", len(c.rooms)))
	return nil
}

func (c *Cache) RoomsInMaintenance(ctx context.Context) ([]string, error) {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()

	var rooms []string
	for room, enabled := range c.rooms {
		if enabled {
			rooms = append(rooms, room)
		}
	}

	return rooms, nil
}

func (c *Cache) RoomInMaintenance(ctx context.Context, room string) (bool, error) {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()
	return c.rooms[room], nil
}

func (c *Cache) SetRoomInMaintenance(ctx context.Context, room string, enabled bool) error {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()

	if c.MaintenanceStore != nil {
		if err := c.MaintenanceStore.SetRoomInMaintenance(ctx, room, enabled); err != nil {
			return fmt.Errorf("unable to set room in maintenance on substore: %w", err)
		}

		c.rooms[room] = enabled
		return nil
	}

	c.rooms[room] = enabled
	return nil
}
