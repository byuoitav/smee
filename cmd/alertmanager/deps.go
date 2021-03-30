package main

import (
	"context"
	"fmt"
	"time"

	"github.com/byuoitav/smee/internal/app/alertmanager/alertcache"
	"github.com/byuoitav/smee/internal/pkg/messenger"
	"github.com/byuoitav/smee/internal/pkg/streamwrapper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (d *Deps) build() {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	d.buildLog()
	d.buildAlertStore(ctx)
	d.buildEventStreamer()
}

func (d *Deps) cleanup() {
	d.log.Sync() // nolint:errcheck
}

func (d *Deps) buildAlertStore(ctx context.Context) {
	store, err := alertcache.New(ctx, nil)
	if err != nil {
		d.log.Fatal("unable to build alert cache", zap.Error(err))
	}

	d.alertStore = store
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
