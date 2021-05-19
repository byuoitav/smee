package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/gin-gonic/gin"
)

func (h *Handlers) MaintenanceInfo(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	maint, err := h.MaintenanceStore.RoomMaintenanceInfo(ctx, c.Param("roomID"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, maint)
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

	c.JSON(http.StatusOK, maint)
}
