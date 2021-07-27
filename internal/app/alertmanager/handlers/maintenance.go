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
	Note   string     `json:"note"`
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
		Note:   info.Note,
	}
}

func (h *Handlers) RoomMaintenanceInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	maint, err := h.MaintenanceStore.RoomMaintenanceInfo(ctx, c.Param("roomID"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, convertMaintenance(maint))
}

func (h *Handlers) SetMaintenanceInfo(c *gin.Context) {
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

func (h *Handlers) RoomsInMaintenance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rooms, err := h.MaintenanceStore.RoomsInMaintenance(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get rooms in maintenance: %s", err)
		return
	}

	c.JSON(http.StatusOK, rooms)
}
