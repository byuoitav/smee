package main

import (
	"context"
	"fmt"
	"os"

	"github.com/byuoitav/smee/internal/app/alertmanager"
	"github.com/byuoitav/smee/internal/app/alertmanager/issuemanager"
	"github.com/byuoitav/smee/internal/smee"
)

func main() {
	alertStore := &struct{}{}
	issueStore := &struct{}{}
	eventStreamer := &struct{}{}
	deviceStateStore := &struct{}{}
	alertManager := &alertmanager.Manager{
		AlertStore:       alertStore,
		IssueStore:       issueStore,
		EventStreamer:    eventStreamer,
		DeviceStateStore: deviceStateStore,
		AlertConfigs:     make(map[string]smee.AlertConfig),
	}
	issueManager := &issuemanager.Manager{
		AlertStore: alertStore,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := alertManager.Run(ctx); err != nil {
		fmt.Printf("unable to run alert manager: %s\n", err)
		os.Exit(1)
	}
}
