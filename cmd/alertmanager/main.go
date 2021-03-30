package main

import (
	"context"
	"fmt"
	"sync"
	"time"

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
	log           *zap.Logger
	alertStore    smee.AlertStore
	alertManager  smee.AlertManager
	eventStreamer smee.EventStreamer
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

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := deps.eventStreamer.Stream(ctx)
		if err != nil {
			deps.log.Fatal("unable to get stream", zap.Error(err))
		}
		defer stream.Close()

		got := 0
		for {
			event, err := stream.Next(ctx)
			if err != nil {
				deps.log.Warn("unable to get event", zap.Error(err))
				break
			}

			got++
			deps.log.Info("Got event", zap.Any("event", event))
		}

		fmt.Printf("1 got: %v\n", got)
	}()

	go func() {
		defer wg.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stream, err := deps.eventStreamer.Stream(ctx)
		if err != nil {
			deps.log.Fatal("unable to get stream", zap.Error(err))
		}
		defer stream.Close()

		got := 0
		for {
			event, err := stream.Next(ctx)
			if err != nil {
				deps.log.Warn("unable to get event", zap.Error(err))
				break
			}

			got++
			deps.log.Info("Got event", zap.Any("event", event))
		}

		fmt.Printf("2 got: %v\n", got)
	}()

	wg.Wait()
	time.Sleep(3 * time.Second)

	/*
		if err := deps.alertManager.Run(context.Background()); err != nil {
			fmt.Printf("unable to run alert manager: %s\n", err)
			os.Exit(1)
		}
	*/
}
