package events

import (
	"context"
	"log/slog"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestBusPublishSubscribe(t *testing.T) {
	bus := NewBus(testLogger())

	var called atomic.Int32
	bus.Subscribe("test.event", func(_ context.Context, e Event) error {
		called.Add(1)
		if e.Type != "test.event" {
			t.Errorf("expected type test.event, got %s", e.Type)
		}
		return nil
	})

	bus.Publish(context.Background(), Event{Type: "test.event", Payload: "hello"})

	time.Sleep(50 * time.Millisecond)
	if called.Load() != 1 {
		t.Errorf("expected handler called 1 time, got %d", called.Load())
	}
}

func TestBusMultipleSubscribers(t *testing.T) {
	bus := NewBus(testLogger())

	var count atomic.Int32
	for range 3 {
		bus.Subscribe("multi", func(_ context.Context, _ Event) error {
			count.Add(1)
			return nil
		})
	}

	bus.Publish(context.Background(), Event{Type: "multi"})
	time.Sleep(50 * time.Millisecond)

	if count.Load() != 3 {
		t.Errorf("expected 3 handlers called, got %d", count.Load())
	}
}

func TestBusNoSubscribers(t *testing.T) {
	bus := NewBus(testLogger())
	// Should not panic
	bus.Publish(context.Background(), Event{Type: "no.sub"})
}
