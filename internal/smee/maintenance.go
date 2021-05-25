package smee

import (
	"context"
	"time"
)

type MaintenanceStore interface {
	// returns a map[roomID]MaintenanceInfo
	RoomsInMaintenance(context.Context) (map[string]MaintenanceInfo, error)
	RoomMaintenanceInfo(ctx context.Context, roomID string) (MaintenanceInfo, error)
	SetMaintenanceInfo(context.Context, MaintenanceInfo) error
}

type MaintenanceInfo struct {
	RoomID string    `json:"roomID"`
	Start  time.Time `json:"start"`
	End    time.Time `json:"end"`
}

func (i MaintenanceInfo) Enabled() bool {
	if i.Start.IsZero() || i.End.IsZero() {
		return false
	}

	now := time.Now()
	return now.After(i.Start) && now.Before(i.End)
}
