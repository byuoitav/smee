package main

import (
	"context"
	"fmt"
	"regexp"
	"time"

	"github.com/byuoitav/auth/wso2"
	"github.com/byuoitav/smee/internal/app/alertmanager"
	"github.com/byuoitav/smee/internal/app/alertmanager/incidents"
	"github.com/byuoitav/smee/internal/app/alertmanager/issuecache"
	"github.com/byuoitav/smee/internal/pkg/messenger"
	"github.com/byuoitav/smee/internal/pkg/servicenow"
	"github.com/byuoitav/smee/internal/pkg/streamwrapper"
	"github.com/byuoitav/smee/internal/smee"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (d *Deps) build() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	d.buildLog()
	d.buildWSO2()
	d.buildIncidentStore()
	d.buildIssueStore(ctx)
	d.buildEventStreamer()
	d.buildAlertManager()
	d.buildHTTPServer(ctx)
}

func (d *Deps) cleanup() {
	d.log.Sync() // nolint:errcheck
}

func (d *Deps) buildIssueStore(ctx context.Context) {
	cache := &issuecache.Cache{
		Log:           d.log.Named("issue-cache"),
		IncidentStore: d.incidentStore,
	}

	if err := cache.Sync(ctx); err != nil {
		d.log.Fatal("unable to sync issue cache", zap.Error(err))
	}

	d.issueStore = cache
}

func (d *Deps) buildIncidentStore() {
	d.incidentStore = &incidents.Store{
		Client: &servicenow.Client{
			Client: d.wso2,
			Log:    d.log.Named("incidents"),
		},
		AssignmentGroup: "OIT-AV Engineers",
		Service:         "TEC Room",
		Priority:        "4",
	}
}

func (d *Deps) buildEventStreamer() {
	if d.HubURL == "" {
		d.log.Fatal("invalid hub url")
	}

	d.eventStreamer = &streamwrapper.StreamWrapper{
		EventStreamer: &messenger.Messenger{
			HubURL: d.HubURL,
		},
	}
}

func (d *Deps) buildAlertManager() {
	d.alertManager = &alertmanager.Manager{
		IssueStore:    d.issueStore,
		EventStreamer: d.eventStreamer,
		AlertConfigs: map[string]smee.AlertConfig{
			"websocket": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:        regexp.MustCompile("^websocket-count$"),
						ValueDoesNotMatch: regexp.MustCompile("^[1-9]{1}[0-9]{0,3}$"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("^websocket-count$"),
						ValueMatches: regexp.MustCompile("^[1-9]{1}[0-9]{0,3}$"),
					},
				},
			},
		},
		Log: d.log.Named("alert-manager"),
	}
}

func (d *Deps) buildWSO2() {
	d.wso2 = wso2.New(d.ClientID, d.ClientSecret, d.GatewayURL, "")
}

func (d *Deps) buildLog() {
	var level zapcore.Level
	if err := level.Set(d.LogLevel); err != nil {
		panic(fmt.Sprintf("invalid log level: %s", err))
	}

	config := zap.Config{
		Level: zap.NewAtomicLevelAt(level),
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding: "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "@",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "trace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	log, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("unable to build logger: %s", err))
	}

	d.log = log
}
