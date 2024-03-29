package main

import (
	"context"
	"fmt"
	"net"

	"github.com/byuoitav/auth/wso2"
	"github.com/byuoitav/smee/internal/app/alertmanager/handlers"
	"github.com/byuoitav/smee/internal/app/commandcli"
	"github.com/byuoitav/smee/internal/pkg/couch"
	"github.com/byuoitav/smee/internal/pkg/postgres"
	"github.com/byuoitav/smee/internal/smee"
	"github.com/byuoitav/smee/opa"
	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type Deps struct {
	// set by command line flags
	Port                 int
	HubURL               string
	LogLevel             string
	ClientID             string
	ClientSecret         string
	GatewayURL           string
	RedirectURL          string
	OPAURL               string
	OPAToken             string
	RedisURL             string
	PostgresURL          string
	DisableAlertManager  bool
	CommandServerAddress string
	CommandToken         string
	CouchURL             string
	CouchUsername        string
	CouchPassword        string
	WebRoot              string

	// created by functions
	log              *zap.Logger
	wso2             *wso2.Client
	opa              *opa.Client
	disableAuth      bool
	postgres         *postgres.Client
	issueStore       smee.IssueStore
	incidentStore    smee.IncidentStore
	maintenanceStore smee.MaintenanceStore
	issuetypeStore   smee.IssueTypeStore
	alertManager     smee.AlertManager
	eventStreamer    smee.EventStreamer
	deviceStateStore smee.DeviceStateStore
	commandClient    *commandcli.Client
	couchManager     *couch.CouchManager

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
	pflag.StringVar(&deps.RedirectURL, "redirect-url", "https://localhost:8080", "wso2 redirect address")
	pflag.StringVar(&deps.OPAURL, "opa-url", "", "the URL for the OPA server to be used for authz")
	pflag.StringVar(&deps.OPAToken, "opa-token", "", "the token to use for calls to OPA")
	pflag.BoolVar(&deps.disableAuth, "disable-auth", false, "disables authz/n checks")
	pflag.StringVar(&deps.RedisURL, "redis-url", "", "redis url")
	pflag.StringVar(&deps.PostgresURL, "postgres-url", "", "postgres url")
	pflag.BoolVar(&deps.DisableAlertManager, "disable-alert-manager", false, "Disables the Alert Management portion of smee")
	pflag.StringVar(&deps.CommandServerAddress, "command-server", "", "url for the av-cli command server")
	pflag.StringVar(&deps.CommandToken, "command-token", "", "the token to use for calls to the av-cli command server")
	pflag.StringVar(&deps.CouchURL, "couch-address", "", "")
	pflag.StringVar(&deps.CouchUsername, "couch-username", "", "")
	pflag.StringVar(&deps.CouchPassword, "couch-password", "", "")
	pflag.StringVar(&deps.WebRoot, "web-root", "/website", "The location on the filesystem of the root of the website files")
	pflag.Parse()

	deps.build()
	defer deps.cleanup()

	g, ctx := errgroup.WithContext(context.Background())

	// Skip turning on the alert manager if we have disabled it
	if !deps.DisableAlertManager {
		g.Go(func() error {
			if err := deps.alertManager.Run(ctx); err != nil {
				return fmt.Errorf("unable to run alert manager: %w", err)
			}

			return fmt.Errorf("alert manager stopped running")
		})
	}

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
