package main

import (
	"context"
	"fmt"
	"net"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Deps struct {
	// set by command line flags
	Port     int
	HubURL   string
	LogLevel string

	// created by functions
	log           *zap.Logger
	issueStore    smee.IssueStore
	alertManager  smee.AlertManager
	eventStreamer smee.EventStreamer

	httpServer   *gin.Engine
	httpListener net.Listener
}

func main() {
	var deps Deps

	pflag.IntVarP(&deps.Port, "port", "P", 8080, "port to run the server on")
	pflag.StringVarP(&deps.LogLevel, "log-level", "L", "", "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVar(&deps.HubURL, "hub-url", "", "url of the event hub")
	pflag.Parse()

	deps.build()
	defer deps.cleanup()

	g, ctx := errgroup.WithContext(context.Background())

	g.Go(func() error {
		if err := deps.alertManager.Run(ctx); err != nil {
			return fmt.Errorf("unable to run alert manager: %w", err)
		}

		return fmt.Errorf("alert manager stopped running")
	})

	g.Go(func() error {
		if err := deps.httpServer.RunListener(deps.httpListener); err != nil {
			return fmt.Errorf("unable to run http server: %w", err)
		}

		return fmt.Errorf("http server stopped running")
	})

	g.Go(func() error {
		<-ctx.Done()
		if err := deps.httpListener.Close(); err != nil {
			return fmt.Errorf("unable to close http listener: %w", err)
		}

		return fmt.Errorf("closed http listener: %w", ctx.Err())
	})

	if err := g.Wait(); err != nil {
		deps.log.Fatal(err.Error())
	}
}
