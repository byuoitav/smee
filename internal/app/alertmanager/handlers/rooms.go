package handlers

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

type room struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	InMaintenance bool   `json:"inMaintenance"`
}

func (h *Handlers) Rooms(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	// TODO get real rooms
	rooms := []room{
		{
			ID:   "ITB-1010",
			Name: "ITB 1010",
		},
		{
			ID:   "ITB-1006",
			Name: "ITB 1006",
		},
		{
			ID:   "ITB-1004",
			Name: "ITB 1004",
		},
		{
			ID:   "ITB-1106",
			Name: "ITB 1106",
		},
		{
			ID:   "JRCB-296",
			Name: "JRCB 296",
		},
	}

	sort.Slice(rooms, func(i, j int) bool {
		return rooms[i].Name < rooms[j].Name
	})

	// merge with maintenance
	maint, err := h.MaintenanceStore.RoomsInMaintenance(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get rooms in maintenance: %s", err)
		return
	}

	for i := range rooms {
		info, ok := maint[rooms[i].ID]
		if ok && info.Enabled() {
			rooms[i].InMaintenance = true
		}
	}

	c.JSON(http.StatusOK, rooms)
}
