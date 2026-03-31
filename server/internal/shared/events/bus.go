package events

import (
	"context"
	"log/slog"
	"sync"
)

type Event struct {
	Type    string
	Payload interface{}
}

type Handler func(ctx context.Context, event Event) error

type Bus interface {
	Publish(ctx context.Context, event Event)
	Subscribe(eventType string, handler Handler)
	Shutdown()
}

type inMemoryBus struct {
	mu       sync.RWMutex
	handlers map[string][]Handler
	logger   *slog.Logger
	wg       sync.WaitGroup
}

func NewBus(logger *slog.Logger) Bus {
	return &inMemoryBus{
		handlers: make(map[string][]Handler),
		logger:   logger,
	}
}

func (b *inMemoryBus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers, ok := b.handlers[event.Type]
	b.mu.RUnlock()

	if !ok {
		return
	}

	for _, h := range handlers {
		b.wg.Add(1)
		go func(handler Handler) {
			defer b.wg.Done()
			if err := handler(ctx, event); err != nil {
				b.logger.Error("event handler failed",
					slog.String("event_type", event.Type),
					slog.String("error", err.Error()),
				)
			}
		}(h)
	}
}

func (b *inMemoryBus) Shutdown() {
	b.wg.Wait()
	b.logger.Info("event bus shutdown complete")
}

func (b *inMemoryBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}
