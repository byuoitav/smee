package main

import (
	"context"
	"fmt"
	"net"

	"github.com/byuoitav/auth/wso2"
	"github.com/byuoitav/smee/internal/app/alertmanager/handlers"
	"github.com/byuoitav/smee/internal/smee"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Deps struct {
	// set by command line flags
	Port         int
	HubURL       string
	LogLevel     string
	ClientID     string
	ClientSecret string
	GatewayURL   string
	RedisURL     string

	// created by functions
	log              *zap.Logger
	wso2             *wso2.Client
	issueStore       smee.IssueStore
	incidentStore    smee.IncidentStore
	maintenanceStore smee.MaintenanceStore
	alertManager     smee.AlertManager
	eventStreamer    smee.EventStreamer
	deviceStateStore smee.DeviceStateStore

	httpServer   *gin.Engine
	handlers     *handlers.Handlers
	httpListener net.Listener
}

func main() {
	var deps Deps

	pflag.IntVarP(&deps.Port, "port", "P", 8080, "port to run the server on")
	pflag.StringVarP(&deps.LogLevel, "log-level", "L", "", "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVar(&deps.HubURL, "hub-url", "", "url of the event hub")
	pflag.StringVar(&deps.ClientID, "client-id", "", "wso2 key")
	pflag.StringVar(&deps.ClientSecret, "client-secret", "", "wso2 secret")
	pflag.StringVar(&deps.GatewayURL, "gateway-url", "https://api.byu.edu", "wso2 gateway address")
	pflag.StringVar(&deps.RedisURL, "redis-url", "", "redis url")
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
