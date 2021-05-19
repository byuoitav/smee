package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/gin-gonic/gin"
)

type maintenanceInfo struct {
	RoomID string     `json:"roomID"`
	Start  *time.Time `json:"start,omitempty"`
	End    *time.Time `json:"end,omitempty"`
}

func convertMaintenance(info smee.MaintenanceInfo) maintenanceInfo {
	var start, end *time.Time

	if !info.Start.IsZero() {
		start = &info.Start
	}

	if !info.End.IsZero() {
		end = &info.End
	}

	return maintenanceInfo{
		RoomID: info.RoomID,
		Start:  start,
		End:    end,
	}
}

func (h *Handlers) MaintenanceInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	maint, err := h.MaintenanceStore.RoomMaintenanceInfo(ctx, c.Param("roomID"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, convertMaintenance(maint))
}

func (h *Handlers) SetRoomInMaintenance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	var maint smee.MaintenanceInfo
	if err := c.Bind(&maint); err != nil {
		c.String(http.StatusBadRequest, "unable to bind: %s", err)
		return
	}

	maint.RoomID = c.Param("roomID")

	if err := h.MaintenanceStore.SetMaintenanceInfo(ctx, maint); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	c.JSON(http.StatusOK, convertMaintenance(maint))
}
