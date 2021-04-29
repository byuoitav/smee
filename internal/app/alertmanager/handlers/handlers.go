package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/gin-gonic/gin"
)

// TODO something to view queue sizes

type Handlers struct {
	IssueStore    smee.IssueStore
	IncidentStore smee.IncidentStore
}

func (h *Handlers) ActiveIssues(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
	defer cancel()

	issues, err := h.IssueStore.ActiveIssues(ctx)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, issues)
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
