package main

import (
	"context"
	"fmt"
	"os"

	"github.com/byuoitav/smee/internal/smee"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

type Deps struct {
	// set by command line flags
	Port     int
	HubURL   string
	LogLevel string

	// created by functions
	log          *zap.Logger
	alertStore   smee.AlertStore
	alertManager smee.AlertManager
}

func main() {
	var deps Deps

	pflag.IntVarP(&deps.Port, "port", "P", 8080, "port to run the server on")
	pflag.StringVarP(&deps.LogLevel, "log-level", "L", "", "level to log at. refer to https://godoc.org/go.uber.org/zap/zapcore#Level for options")
	pflag.StringVar(&deps.HubURL, "hub-url", "", "url of the event hub")
	pflag.Parse()

	deps.build()
	defer deps.cleanup()

	// issueStore := &struct{}{}
	// eventStreamer := &struct{}{}
	// deviceStateStore := &struct{}{}
	/*
		alertManager := &alertmanager.Manager{
			AlertStore:       alertStore,
			IssueStore:       issueStore,
			EventStreamer:    eventStreamer,
			DeviceStateStore: deviceStateStore,
			AlertConfigs:     make(map[string]smee.AlertConfig),
		}
	*/

	if err := deps.alertManager.Run(context.Background()); err != nil {
		fmt.Printf("unable to run alert manager: %s\n", err)
		os.Exit(1)
	}
}
