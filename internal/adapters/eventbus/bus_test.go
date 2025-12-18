package eventbus

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishCallsHandler(t *testing.T) {
	b := NewInMemoryBus(1, 10)
	defer b.Close()

	ch := make(chan interface{}, 1)
	unsub := b.Subscribe("topic1", func(ctx context.Context, payload interface{}) error {
		ch <- payload
		return nil
	})
	defer unsub()

	b.Publish(context.Background(), "topic1", "hello")

	select {
	case v := <-ch:
		if v != "hello" {
			t.Fatalf("expected payload 'hello', got %v", v)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("handler was not called in time")
	}
}

func TestMultipleHandlersReceive(t *testing.T) {
	b := NewInMemoryBus(2, 10)
	defer b.Close()

	ch1 := make(chan interface{}, 1)
	ch2 := make(chan interface{}, 1)

	u1 := b.Subscribe("topic-multi", func(ctx context.Context, payload interface{}) error {
		ch1 <- payload
		return nil
	})
	defer u1()
	u2 := b.Subscribe("topic-multi", func(ctx context.Context, payload interface{}) error {
		ch2 <- payload
		return nil
	})
	defer u2()

	b.Publish(context.Background(), "topic-multi", 42)

	timeout := time.After(500 * time.Millisecond)
	got1, got2 := false, false
	for !(got1 && got2) {
		select {
		case <-ch1:
			got1 = true
		case <-ch2:
			got2 = true
		case <-timeout:
			t.Fatal("not all handlers received the message")
		}
	}
}

func TestUnsubscribeStopsReceiving(t *testing.T) {
	b := NewInMemoryBus(1, 10)
	defer b.Close()

	ch := make(chan interface{}, 1)
	unsub := b.Subscribe("topic-unsub", func(ctx context.Context, payload interface{}) error {
		ch <- payload
		return nil
	})

	// unsubscribe immediately
	unsub()

	b.Publish(context.Background(), "topic-unsub", "x")

	select {
	case v := <-ch:
		t.Fatalf("expected no message after unsubscribe, got %v", v)
	case <-time.After(200 * time.Millisecond):
		// success: no message
	}
}

func TestPublishDoesNotBlockWhenQueueFull(t *testing.T) {
	// small queue and slow handler to force fallback path
	b := NewInMemoryBus(1, 1)
	defer b.Close()

	wait := make(chan struct{})
	var handled int32

	u := b.Subscribe("topic-busy", func(ctx context.Context, payload interface{}) error {
		atomic.AddInt32(&handled, 1)
		// block until released to simulate slow handler
		<-wait
		return nil
	})
	defer u()

	start := time.Now()
	// first publish will occupy the queue/worker
	b.Publish(context.Background(), "topic-busy", 1)
	// second publish should not block even if queue is full (fallback goroutine)
	b.Publish(context.Background(), "topic-busy", 2)
	elapsed := time.Since(start)
	if elapsed > 200*time.Millisecond {
		t.Fatalf("Publish appears to block when queue full (took %v)", elapsed)
	}

	// allow handlers to finish
	close(wait)

	// give some time for handler goroutines to finish
	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&handled) < 2 {
		t.Fatalf("expected at least 2 handled messages, got %d", handled)
	}
}

func TestNewInMemoryBusDefaultsAndPublish(t *testing.T) {
	b := NewInMemoryBus(0, 0) // should use defaults
	defer b.Close()

	ch := make(chan interface{}, 1)
	unsub := b.Subscribe("dft", func(ctx context.Context, payload interface{}) error {
		ch <- payload
		return nil
	})
	defer unsub()

	b.Publish(context.Background(), "dft", "ok")

	select {
	case v := <-ch:
		if v != "ok" {
			t.Fatalf("expected ok, got %v", v)
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("handler not called")
	}
}

func TestPublishWithNoHandlersDoesNotPanic(t *testing.T) {
	b := NewInMemoryBus(1, 1)
	defer b.Close()

	// no subscribers for topic 'none'
	b.Publish(context.Background(), "none", 123)
}

func TestCloseTwiceReturnsError(t *testing.T) {
	b := NewInMemoryBus(1, 1)
	if err := b.Close(); err != nil {
		t.Fatalf("unexpected close error: %v", err)
	}
	if err := b.Close(); err == nil {
		t.Fatalf("expected error on second close")
	}
}

func TestPublishEnqueuedPath(t *testing.T) {
	b := NewInMemoryBus(1, 10)
	defer b.Close()

	var handled int32
	unsub := b.Subscribe("enqueue", func(ctx context.Context, payload interface{}) error {
		atomic.AddInt32(&handled, 1)
		return nil
	})
	defer unsub()

	b.Publish(context.Background(), "enqueue", 1)

	timeout := time.After(500 * time.Millisecond)
	for atomic.LoadInt32(&handled) == 0 {
		select {
		case <-timeout:
			t.Fatal("handler not run via enqueue path")
		default:
			time.Sleep(10 * time.Millisecond)
		}
	}
}
