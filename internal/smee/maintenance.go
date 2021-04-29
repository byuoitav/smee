package smee

type MaintenanceStore struct {
	RoomsInMaintenance(context.Context) ([]string, error)
}
