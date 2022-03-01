package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path"
	"path/filepath"

	"github.com/byuoitav/auth/session/cookiestore"
	"github.com/byuoitav/smee/internal/app/alertmanager/handlers"
	"github.com/gin-gonic/gin"
	adapter "github.com/gwatts/gin-adapter"
	"go.uber.org/zap"
)

func (d *Deps) buildHTTPServer(ctx context.Context) {
	// build listener
	var lc net.ListenConfig
	var err error

	d.httpListener, err = lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", d.Port))
	if err != nil {
		d.log.Fatal("unable to bind listener", zap.Error(err))
	}

	d.handlers = &handlers.Handlers{
		IssueStore:       d.issueStore,
		MaintenanceStore: d.maintenanceStore,
		IncidentStore:    d.incidentStore,
		IssueTypeStore:   d.issuetypeStore,
	}

	// build engine
	r := gin.New()
	r.Use(gin.Recovery())

	sessionStore := cookiestore.NewStore()

	//auth
	if !d.disableAuth {
		if d.opa.URL == "" {
			d.log.Fatal("No OPA URL was set, but authz has not been disabled")
		}
		r.Use(adapter.Wrap(d.wso2.AuthCodeMiddleware(sessionStore, "smee")))
		r.Use(d.opa.Authorize())
	}

	r.NoRoute(func(c *gin.Context) {
		dir, file := path.Split(c.Request.RequestURI)

		if file == "" || filepath.Ext(file) == "" {
			c.File(fmt.Sprintf("%s/index.html", d.WebRoot))
		} else {
			c.File(fmt.Sprintf("%s/", d.WebRoot) + path.Join(dir, file))
		}
	})

	debug := r.Group("/debug")
	debug.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	api := r.Group("api/v1")

	api.GET("/issues", d.handlers.ActiveIssues)
	api.PUT("/issues/:issueID/linkIncident", d.handlers.LinkIssueToIncident)
	api.PUT("/issues/:issueID/createIncident", d.handlers.CreateIncidentFromIssue)
	api.PUT("/issues/:issueID/closeIssue", d.handlers.CloseIssue)
	api.PUT("/issues/:issueID/acknowledgeIssue", d.handlers.AcknowledgeIssue)
	api.PUT("/issues/:issueID/unacknowledgeIssue", d.handlers.UnacknowledgeIssue)
	api.PUT("/issues/:issueID/setStatus", d.handlers.SetStatus)

	api.GET("/maintenance", d.handlers.RoomsInMaintenance)
	api.GET("/maintenance/:roomID", d.handlers.RoomMaintenanceInfo)
	api.PUT("/maintenance/:roomID", d.handlers.SetMaintenanceInfo)

	api.GET("/rooms", d.handlers.Rooms)
	api.GET("/issuetype", d.handlers.SNIssueType)

	d.httpServer = r
}
