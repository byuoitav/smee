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
	IssueTypeStore   smee.IssueTypeStore
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

type IssType struct {
	Type map[string]smee.IssueType `json:"IssueType"`
}

func (h *Handlers) SNIssueType(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	issueT, err := h.IssueTypeStore.IssueType(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to get Service Now Alert Type: %s", err)
		return
	}

	var issT IssType
	issT.Type = issueT

	c.JSON(http.StatusOK, issT)
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

		c.JSON(http.StatusOK, convertIssue(issue))
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
		issue := convertIssue(iss)

		issue.MaintenanceStart = info.Start
		issue.MaintenanceEnd = info.End

		res = append(res, issue)
	}

	c.JSON(http.StatusOK, res)
}

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

func (h *Handlers) CloseIssue(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	issueID := c.Param("issueID")
	iss, err := h.IssueStore.CloseAlertsForIssue(ctx, issueID)
	if err != nil {
		c.String(http.StatusInternalServerError, "unable to close issue: %s", err)
		return
	}

	c.JSON(http.StatusOK, iss)
}

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
		Caller:           "avmonit1", // using av monitoring user as caller
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

func convertIssue(iss smee.Issue) issue {
	issue := issue{
		ID:        iss.ID,
		Room:      iss.Room,
		Start:     iss.Start,
		Alerts:    make(map[string]alert, len(iss.Alerts)),
		Incidents: iss.Incidents,
		Events:    make([]issueEvent, len(iss.Events)),
	}

	if !iss.End.IsZero() {
		issue.End = &iss.End
	}

	for i, event := range iss.Events {
		tempData := event.Data

		issue.Events[i] = issueEvent{
			Timestamp: event.Timestamp,
			Type:      string(event.Type),
			Data:      &tempData, // TODO should i make a new slice and copy()?
		}
	}

	for i := range iss.Alerts {
		alert := alert{
			ID:      iss.Alerts[i].ID,
			IssueID: iss.Alerts[i].IssueID,
			Device:  iss.Alerts[i].Device,
			Type:    iss.Alerts[i].Type,
			Start:   iss.Alerts[i].Start,
		}

		if !iss.Alerts[i].End.IsZero() {
			tempEnd := iss.Alerts[i].End
			alert.End = &tempEnd
		}
		issue.Alerts[alert.ID] = alert

	}
	return issue
}
