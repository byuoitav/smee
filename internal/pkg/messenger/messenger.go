package messenger

import (
	"context"
	"errors"
	"fmt"

	"github.com/byuoitav/central-event-system/hub/base"
	"github.com/byuoitav/central-event-system/messenger"
	"github.com/byuoitav/common/v2/events"
	"github.com/byuoitav/smee/internal/smee"
)

type Messenger struct {
	HubURL string
}

type stream struct {
	m    *messenger.Messenger
	done chan struct{}
}

func (m *Messenger) Stream(ctx context.Context) (smee.EventStream, error) {
	mess, err := messenger.BuildMessenger(m.HubURL, base.Messenger, 10000)
	if err != nil {
		return nil, fmt.Errorf("unable to build messenger: %w", err)
	}

	mess.SubscribeToRooms("*")
	return &stream{
		m:    mess,
		done: make(chan struct{}),
	}, nil
}

func (s *stream) Next(ctx context.Context) (smee.Event, error) {
	event := make(chan events.Event)
	go func() {
		event <- s.m.ReceiveEvent()
	}()

	select {
	case event := <-event:
		return smee.Event{
			RoomID:   event.TargetDevice.RoomID,
			DeviceID: event.TargetDevice.DeviceID,
			Key:      event.Key,
			Value:    event.Value,
		}, nil
	case <-s.done:
		return smee.Event{}, errors.New("stream closed")
	case <-ctx.Done():
		return smee.Event{}, ctx.Err()
	}
}

func (s *stream) Close() error {
	close(s.done)
	return nil
}
