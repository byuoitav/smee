package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/gin-gonic/gin"
)

// TODO something to view queue sizes

type Handlers struct {
	IssueStore       smee.IssueStore
	IncidentStore    smee.IncidentStore
	MaintenanceStore smee.MaintenanceStore
}

type issue struct {
	ID               string                   `json:"id"`
	Room             smee.Room                `json:"room"`
	Start            time.Time                `json:"start"`
	End              *time.Time               `json:"end,omitempty"`
	Alerts           map[string]alert         `json:"alerts"`
	Incidents        map[string]smee.Incident `json:"incidents"`
	Events           []issueEvent             `json:"events"`
	MaintenanceStart *time.Time               `json:"maintenanceStart,omitempty"`
	MaintenanceEnd   *time.Time               `json:"maintenanceEnd,omitempty"`
}

type alert struct {
	ID      string      `json:"id"`
	IssueID string      `json:"issueID"`
	Device  smee.Device `json:"device"`
	Type    string      `json:"type"`
	Start   time.Time   `json:"start"`
	End     *time.Time  `json:"end"`
}

type issueEvent struct {
	Timestamp time.Time        `json:"timestamp"`
	Type      string           `json:"type"`
	Data      *json.RawMessage `json:"data"`
}

func (h *Handlers) ActiveIssues(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	roomID := c.Query("roomID")
	if len(roomID) > 0 {
		// get issue for this room
		issue, err := h.IssueStore.ActiveIssue(ctx, roomID)
		switch {
		case err != nil:
			c.String(http.StatusInternalServerError, err.Error())
			return
		case errors.Is(err, smee.ErrRoomIssueNotFound):
			c.Status(http.StatusNotFound)
			return
		}

		c.JSON(http.StatusOK, issue)
		return
	}

	// get all issues
	issues, err := h.IssueStore.ActiveIssues(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get active issues: %s", err)
		return
	}

	// get maintenance info
	maint, err := h.MaintenanceStore.RoomsInMaintenance(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get maintenance info: %s", err)
		return
	}

	var res []issue
	for _, iss := range issues {
		info := convertMaintenance(maint[iss.Room.ID])
		issue := issue{
			ID:               iss.ID,
			Room:             iss.Room,
			Start:            iss.Start,
			Alerts:           make(map[string]alert, len(iss.Alerts)),
			Incidents:        iss.Incidents,
			Events:           make([]issueEvent, len(iss.Events)),
			MaintenanceStart: info.Start,
			MaintenanceEnd:   info.End,
		}

		if !iss.End.IsZero() {
			issue.End = &iss.End
		}

		for i, event := range iss.Events {
			issue.Events[i] = issueEvent{
				Timestamp: event.Timestamp,
				Type:      string(event.Type),
				Data:      &event.Data,
			}
		}

		for _, a := range iss.Alerts {
			alert := alert{
				ID:      a.ID,
				IssueID: a.IssueID,
				Device:  a.Device,
				Type:    a.Type,
				Start:   a.Start,
			}

			if !a.End.IsZero() {
				alert.End = &a.End
			}

			issue.Alerts[alert.ID] = alert
		}

		res = append(res, issue)
	}

	c.JSON(http.StatusOK, res)
}

// TODO maintenance
func (h *Handlers) LinkIssueToIncident(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	issueID := c.Param("issueID")
	incName := c.Query("incName")
	if len(incName) == 0 {
		c.String(http.StatusBadRequest, "must include incName")
		return
	}

	inc, err := h.IncidentStore.IncidentByName(ctx, incName)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get incident: %s", err)
		return
	}

	iss, err := h.IssueStore.LinkIncident(ctx, issueID, inc)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to link incident: %s", err)
		return
	}

	c.JSON(http.StatusOK, iss)
}

// TODO maintenance
func (h *Handlers) CreateIncidentFromIssue(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	issueID := c.Param("issueID")
	shortDesc := c.Query("shortDescription")
	if len(shortDesc) == 0 {
		c.String(http.StatusBadRequest, "must include shortDescription")
		return
	}

	inc := smee.Incident{
		ShortDescription: shortDesc,
		Caller:           "", // pull from context once auth is done
	}

	inc, err := h.IncidentStore.CreateIncident(ctx, inc)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to create incident: %s", err)
		return
	}

	iss, err := h.IssueStore.LinkIncident(ctx, issueID, inc)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to link incident: %s", err)
		return
	}

	c.JSON(http.StatusOK, iss)
}
