package smee

import "context"

type MaintenanceStore interface {
	RoomsInMaintenance(context.Context) ([]string, error)
	RoomInMaintenance(context.Context, string) (bool, error)
	SetRoomInMaintenance(context.Context, string, bool) error
}
