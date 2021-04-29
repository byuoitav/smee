package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handlers) RoomsInMaintenance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rooms, err := h.MaintenanceStore.RoomsInMaintenance(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, rooms)
}

func (h *Handlers) RoomInMaintenance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	maint, err := h.MaintenanceStore.RoomInMaintenance(ctx, c.Param("roomID"))
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"maintenance": maint,
	})
}

func (h *Handlers) SetRoomInMaintenance(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	maint, err := strconv.ParseBool(c.Query("maintenance"))
	if err != nil {
		c.String(http.StatusBadRequest, "must include maintenance as a bool")
		return
	}

	if err := h.MaintenanceStore.SetRoomInMaintenance(ctx, c.Param("roomID"), maint); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	c.Status(http.StatusOK)
}
