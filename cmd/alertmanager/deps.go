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
	"github.com/byuoitav/smee/internal/app/alertmanager/maintenance"
	"github.com/byuoitav/smee/internal/app/alertmanager/redis"
	"github.com/byuoitav/smee/internal/pkg/messenger"
	"github.com/byuoitav/smee/internal/pkg/postgres"
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
	d.buildIncidentMaintenanceStore(ctx)
	d.buildIncidentStore()
	d.buildIssueCache(ctx)
	d.buildMaintenanceCache(ctx)
	d.buildIssueTypeStore(ctx)

	// Disable building alert management stuff if we have disabled it
	if !d.DisableAlertManager {
		d.buildEventStreamer()
		d.buildDeviceStateStore(ctx)
		d.buildAlertManager()
	}

	d.buildHTTPServer(ctx)
}

func (d *Deps) cleanup() {
	d.log.Sync()       // nolint:errcheck
	d.postgres.Close() // nolint:errcheck
}

func (d *Deps) buildIncidentMaintenanceStore(ctx context.Context) {
	store, err := postgres.New(ctx, d.PostgresURL)
	if err != nil {
		d.log.Fatal("unable to build postgres store", zap.Error(err))
	}

	store.Log = d.log.Named("postgres")

	d.postgres = store
	d.issueStore = store
	d.maintenanceStore = store
}

func (d *Deps) buildIssueTypeStore(ctx context.Context) {
	d.issuetypeStore = d.postgres
}

func (d *Deps) buildIssueCache(ctx context.Context) {
	cache := &issuecache.Cache{
		Log:           d.log.Named("issue-cache"),
		IncidentStore: d.incidentStore,
		IssueStore:    d.issueStore,
	}

	if err := cache.Sync(ctx); err != nil {
		d.log.Fatal("unable to sync issue cache", zap.Error(err))
	}

	d.issueStore = cache
}

func (d *Deps) buildMaintenanceCache(ctx context.Context) {
	cache := &maintenance.Cache{
		Log:              d.log.Named("maintenance-cache"),
		MaintenanceStore: d.maintenanceStore,
	}

	if err := cache.Sync(ctx); err != nil {
		d.log.Fatal("unable to sync maintenance cache", zap.Error(err))
	}

	d.maintenanceStore = cache
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

func (d *Deps) buildDeviceStateStore(ctx context.Context) {
	store, err := redis.New(ctx, d.RedisURL)
	if err != nil {
		d.log.Fatal("unable to build redis store", zap.Error(err))
	}

	d.deviceStateStore = store
}

func (d *Deps) buildAlertManager() {
	d.alertManager = &alertmanager.Manager{
		IssueStore:       d.issueStore,
		MaintenanceStore: d.maintenanceStore,
		EventStreamer:    d.eventStreamer,
		DeviceStateStore: d.deviceStateStore,
		AlertConfigs: map[string]smee.AlertConfig{
			"cpu-temperature": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("thermal0-temp"),
						ValueMatches: regexp.MustCompile(`^([8-9][0-9]|[1-9][0-9]{2,})(\.[0-9]*)*$`),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("thermal0-temp"),
						ValueMatches: regexp.MustCompile(`^0*([0-9]|[1-6][0-9])(\.[0-9]*)*$`),
					},
				},
			},
			"device-comm": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:        regexp.MustCompile("^responsive$"),
						ValueDoesNotMatch: regexp.MustCompile("^Ok$"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("^responsive$"),
						ValueMatches: regexp.MustCompile("^Ok$"),
					},
				},
			},
			"device-offline": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:        regexp.MustCompile("^online$"),
						ValueDoesNotMatch: regexp.MustCompile("^Online$"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("^online$"),
						ValueMatches: regexp.MustCompile("^Online$"),
					},
				},
			},
			"lamp-warning": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("status-message"),
						ValueMatches: regexp.MustCompile("(?i)WARNING|Communication|AROUND LAMP TEMPERATURE"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("status-message"),
						ValueMatches: regexp.MustCompile("NO ERRORS|Normal"),
					},
				},
			},
			"memory-usage": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("^v-mem-used-percent$"),
						ValueMatches: regexp.MustCompile("^([9][0-9]|[1-9][0-9]{2,})$"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("^v-mem-used-percent$"),
						ValueMatches: regexp.MustCompile(`^0*([0-9]|[1-8][0-9])\.`),
					},
				},
			},
			"shutter-error": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("status-message"),
						ValueMatches: regexp.MustCompile("SHUTTER ERROR"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("status-message"),
						ValueMatches: regexp.MustCompile("NO ERRORS"),
					},
				},
			},
			"touchpanel-offline": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:        regexp.MustCompile("^tp_online$"),
						ValueDoesNotMatch: regexp.MustCompile("^Online$"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("^tp_online$"),
						ValueMatches: regexp.MustCompile("^Online$"),
					},
				},
			},
			/*
				"interference": { // can't figure out how to do this one...
					Create: smee.AlertTransition{
						Event: &smee.AlertTransitionEvent{
							KeyMatches:   regexp.MustCompile("interference"),
							ValueMatches: regexp.MustCompile("Interference Detected"),
						},
					},
					Close: smee.AlertTransition{
						Event: &smee.AlertTransitionEvent{
							KeyMatches:   regexp.MustCompile("^tp_online$"),
							ValueMatches: regexp.MustCompile("^Online$"),
						},
					},
				},
			*/
			"mic-battery": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("battery-charge-minutes"),
						ValueMatches: regexp.MustCompile("^0*(([1-8][0-9])|([0-9])|90)$"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("battery-charge-minutes"),
						ValueMatches: regexp.MustCompile("^0*(([9][1-9])|[1-9][0-9]{2,})$"),
					},
				},
			},
			"receiver": {
				Create: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:        regexp.MustCompile("mic-alerting"),
						ValueDoesNotMatch: regexp.MustCompile("Okay"),
					},
				},
				Close: smee.AlertTransition{
					Event: &smee.AlertTransitionEvent{
						KeyMatches:   regexp.MustCompile("mic-alerting"),
						ValueMatches: regexp.MustCompile("Okay"),
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
