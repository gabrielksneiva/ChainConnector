package ports

import "context"

type EventHandler func(ctx context.Context, payload interface{}) error

type EventBus interface {
	Publish(ctx context.Context, topic string, payload interface{})
	Subscribe(topic string, handler EventHandler) func()
	Close() error
}
