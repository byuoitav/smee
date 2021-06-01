package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/jackc/pgx/v4"
)

type maintenanceInfo struct {
	CouchRoomID string
	StartTime   time.Time
	EndTime     time.Time
}

func (c *Client) RoomsInMaintenance(ctx context.Context) (map[string]smee.MaintenanceInfo, error) {
	maint := make(map[string]smee.MaintenanceInfo)
	var info maintenanceInfo

	_, err := c.pool.QueryFunc(ctx,
		"SELECT * FROM room_maintenance_couch WHERE now() BETWEEN start_time AND end_time",
		[]interface{}{},
		[]interface{}{&info.CouchRoomID, &info.StartTime, &info.EndTime},
		func(pgx.QueryFuncRow) error {
			smeeInfo := convertMaintenanceInfo(info)
			maint[smeeInfo.RoomID] = smeeInfo
			return nil
		},
	)
	if err != nil {
		return nil, fmt.Errorf("unable to queryFunc: %w", err)
	}

	return maint, nil
}

func (c *Client) RoomMaintenanceInfo(ctx context.Context, roomID string) (smee.MaintenanceInfo, error) {
	var info maintenanceInfo

	err := c.pool.QueryRow(ctx,
		"SELECT * FROM room_maintenance_couch WHERE couch_room_id = $1",
		roomID).Scan(&info.CouchRoomID, &info.StartTime, &info.EndTime)
	switch {
	case err == pgx.ErrNoRows:
		return smee.MaintenanceInfo{}, smee.ErrRoomIssueNotFound // TODO change error type
	case err != nil:
		return smee.MaintenanceInfo{}, fmt.Errorf("unable to query/scan: %w", err)
	}

	return convertMaintenanceInfo(info), nil
}

func (c *Client) SetMaintenanceInfo(ctx context.Context, info smee.MaintenanceInfo) error {
	_, err := c.pool.Exec(ctx,
		"INSERT INTO room_maintenance_couch (couch_room_id, start_time, end_time) values ($1, $2, $3) ON CONFLICT (couch_room_id) DO UPDATE SET start_time = EXCLUDED.start_time, end_time = EXCLUDED.end_time",
		info.RoomID, info.Start, info.End)
	if err != nil {
		return fmt.Errorf("unable to exec :%w", err)
	}

	return nil
}

func convertMaintenanceInfo(info maintenanceInfo) smee.MaintenanceInfo {
	smeeInfo := smee.MaintenanceInfo{
		RoomID: info.CouchRoomID,
		Start:  info.StartTime,
		End:    info.EndTime,
	}

	return smeeInfo
}
