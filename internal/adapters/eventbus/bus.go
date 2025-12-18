package eventbus

import (
	"ChainConnector/internal/domain/ports"
	"context"
	"errors"
	"sync"
)

type job struct {
	handler ports.EventHandler
	payload interface{}
	ctx     context.Context
}

type InMemoryBus struct {
	mu     sync.RWMutex
	subs   map[string]map[int]ports.EventHandler
	nextID int

	jobs       chan job
	workerWG   sync.WaitGroup
	stop       chan struct{}
	stopClosed bool
}

var _ ports.EventBus = (*InMemoryBus)(nil)

func NewInMemoryBus(workerCount, queueSize int) *InMemoryBus {
	if workerCount <= 0 {
		workerCount = 4
	}
	if queueSize <= 0 {
		queueSize = 1024
	}
	b := &InMemoryBus{
		subs: make(map[string]map[int]ports.EventHandler),
		jobs: make(chan job, queueSize),
		stop: make(chan struct{}),
	}
	for i := 0; i < workerCount; i++ {
		b.workerWG.Add(1)
		go func() {
			defer b.workerWG.Done()
			for {
				select {
				case j := <-b.jobs:
					_ = j.handler(j.ctx, j.payload)
				case <-b.stop:
					return
				}
			}
		}()
	}
	return b
}

// Subscribe adiciona handler e retorna função de unsubscribe.
func (b *InMemoryBus) Subscribe(topic string, h ports.EventHandler) func() {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.nextID
	b.nextID++
	if b.subs[topic] == nil {
		b.subs[topic] = make(map[int]ports.EventHandler)
	}
	b.subs[topic][id] = h

	return func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		delete(b.subs[topic], id)
	}
}

// Publish enfileira jobs para os handlers; se a fila estiver cheia, faz fallback para goroutine.
func (b *InMemoryBus) Publish(ctx context.Context, topic string, payload interface{}) {
	b.mu.RLock()
	handlersMap := b.subs[topic]
	b.mu.RUnlock()

	if handlersMap == nil {
		return
	}

	// snapshot handlers
	handlers := make([]ports.EventHandler, 0, len(handlersMap))
	for _, h := range handlersMap {
		handlers = append(handlers, h)
	}

	for _, h := range handlers {
		j := job{handler: h, payload: payload, ctx: ctx}
		select {
		case b.jobs <- j:
			// enqueued
		default:
			// fila cheia: fallback para execução imediata em goroutine (fire-and-forget)
			go func(h ports.EventHandler, ctx context.Context, payload interface{}) {
				_ = h(ctx, payload)
			}(h, ctx, payload)
		}
	}
}

// Close sinaliza o stop e espera workers terminarem.
func (b *InMemoryBus) Close() error {
	b.mu.Lock()
	if b.stopClosed {
		b.mu.Unlock()
		return errors.New("bus already closed")
	}
	b.stopClosed = true
	close(b.stop)
	b.mu.Unlock()

	b.workerWG.Wait()
	return nil
}
