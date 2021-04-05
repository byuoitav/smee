package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (d *Deps) buildHTTPServer() {
	r := gin.New()
	r.Use(gin.Recovery())

	debug := r.Group("/debug")
	debug.GET("/healthz", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	d.httpServer = r
}
