package streamwrapper

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/byuoitav/smee/internal/smee"
)

type StreamWrapper struct {
	EventStreamer smee.EventStreamer

	once      sync.Once
	mu        sync.Mutex
	streaming bool
	streams   map[*wrappedStream]struct{}
}

type wrappedStream struct {
	wrapper *StreamWrapper
	events  chan smee.Event
}

func (s *StreamWrapper) Stream(ctx context.Context) (smee.EventStream, error) {
	s.once.Do(func() {
		s.streams = make(map[*wrappedStream]struct{})
	})

	s.mu.Lock()
	defer s.mu.Unlock()

	wrapped := &wrappedStream{
		wrapper: s,
		// this channel is buffered to make it less likely for
		// events to be missed if a receiver is busy
		events: make(chan smee.Event, 512),
	}
	s.streams[wrapped] = struct{}{}

	if !s.streaming {
		// create a new base stream
		stream, err := s.EventStreamer.Stream(ctx)
		if err != nil {
			return nil, fmt.Errorf("unable to create base stream: %w", err)
		}

		go s.startStream(stream)
		s.streaming = true
	}

	return wrapped, nil
}

func (s *StreamWrapper) startStream(stream smee.EventStream) {
	defer stream.Close()
	defer func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		s.streaming = false
	}()

	next := func() (smee.Event, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		return stream.Next(ctx)
	}
	distribute := func(event smee.Event) bool {
		s.mu.Lock()
		defer s.mu.Unlock()

		for wrapped := range s.streams {
			select {
			case wrapped.events <- event:
			default:
			}
		}

		return len(s.streams) == 0
	}

	for {
		event, err := next()
		if err != nil {
			return
		}

		if done := distribute(event); done {
			return
		}
	}
}

func (s *wrappedStream) Next(ctx context.Context) (smee.Event, error) {
	select {
	case event, ok := <-s.events:
		if !ok {
			return smee.Event{}, errors.New("stream closed")
		}

		return event, nil
	case <-ctx.Done():
		return smee.Event{}, ctx.Err()
	}
}

func (s *wrappedStream) Close() error {
	s.wrapper.mu.Lock()
	defer s.wrapper.mu.Unlock()
	delete(s.wrapper.streams, s)
	close(s.events)
	return nil
}
