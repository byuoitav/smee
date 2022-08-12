package handlers

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type roomOverview struct {
	ID            string `json:"id"`
	InMaintenance bool   `json:"inMaintenance"`
}

func (h *Handlers) Rooms(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	rooms, err := h.CouchManager.GetRooms()
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get rooms in maintenance: %s", err)
		return
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i] < rooms[j]
	})

	// merge with maintenance
	maint, err := h.MaintenanceStore.RoomsInMaintenance(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get rooms in maintenance: %s", err)
		return
	}

	roomList := []roomOverview{}

	for i := range rooms {
		roomList = append(roomList, roomOverview{
			ID:            rooms[i],
			InMaintenance: false,
		})

		info, ok := maint[rooms[i]]
		if ok && info.Enabled() {
			roomList[i].InMaintenance = true
		}
	}

	c.JSON(http.StatusOK, roomList)
}
