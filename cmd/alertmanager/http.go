package main

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
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

	// build engine
	r := gin.New()
	r.Use(gin.Recovery())

	debug := r.Group("/debug")
	debug.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	d.httpServer = r
}
