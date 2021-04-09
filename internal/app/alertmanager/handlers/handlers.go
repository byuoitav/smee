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
	IssueStore smee.IssueStore
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
