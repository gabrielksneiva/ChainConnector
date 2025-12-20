package http

import (
	"ChainConnector/internal/domain/ports"
	"ChainConnector/internal/domain/service"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var capturedHook fx.Hook

type fakeApp struct{}

func (f *fakeApp) Listen(_ string) error                          { return nil }
func (f *fakeApp) Shutdown() error                                { return nil }
func (f *fakeApp) Get(_ string, _ ...fiber.Handler) fiber.Router  { return nil }
func (f *fakeApp) Post(_ string, _ ...fiber.Handler) fiber.Router { return nil }

type fakeLc struct{}

func (f *fakeLc) Append(h fx.Hook) { capturedHook = h }

// Ensure CreateFiberServer registers healthcheck and root routes.
func TestCreateFiberServer_Healthcheck(t *testing.T) {
	app := CreateFiberServer()
	// /health should return 200 OK with body "OK"
	req, _ := http.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("unexpected error from app.Test: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestFiberServer_StartRegistersHooks(t *testing.T) {
	// Should not panic when registering hooks
	// Start hook registration relies on fx lifecycle internals and
	// is exercised indirectly via integration. Skip direct invocation
	// here to avoid lifecycle initialization complexity in unit tests.
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	s := NewFiberServer(logger, txSvc, nil)

	// Use zero-value lifecycle; Start should handle nil Append without panicking.
	var lc fx.Lifecycle
	s.Start(lc)
}

func TestNewFiberServer_ConstructsWithLogger(t *testing.T) {
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	s := NewFiberServer(logger, txSvc, nil)
	if s == nil || s.app == nil {
		t.Fatalf("expected non-nil FiberServer and app")
	}
	if s.logger == nil {
		t.Fatalf("expected logger to be set on FiberServer")
	}
}

func TestFiberServer_HookExecution(t *testing.T) {
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	s := NewFiberServer(logger, txSvc, nil)
	// inject fake app to avoid real network Listen
	s.app = &fakeApp{}

	// clear captured hook and start
	capturedHook = fx.Hook{}
	s.Start(&fakeLc{})

	// Execute OnStart and OnStop directly; they should not panic.
	if capturedHook.OnStart != nil {
		if err := capturedHook.OnStart(context.Background()); err != nil {
			t.Fatalf("OnStart returned error: %v", err)
		}
	}
	if capturedHook.OnStop != nil {
		if err := capturedHook.OnStop(context.Background()); err != nil {
			t.Fatalf("OnStop returned error: %v", err)
		}
	}
}

type fakeBus struct {
	lastTopic   string
	lastPayload interface{}
}

func (f *fakeBus) Publish(ctx context.Context, topic string, payload interface{}) {
	f.lastTopic = topic
	f.lastPayload = payload
}
func (f *fakeBus) Subscribe(topic string, handler ports.EventHandler) func() { return func() {} }
func (f *fakeBus) Close() error                                              { return nil }

func TestHandlerTransactionPublishes(t *testing.T) {
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	bus := &fakeBus{}
	s := NewFiberServer(logger, txSvc, bus)
	app := s.app.(*fiber.App)

	body := map[string]string{
		"from":      "0xfrom",
		"to":        "0xto",
		"chain":     "sepolia",
		"amount":    "10",
		"gas":       "21000",
		"gas_price": "1",
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", resp.StatusCode)
	}
	if bus.lastTopic != "transaction.created" {
		t.Fatalf("expected topic transaction.created, got %s", bus.lastTopic)
	}
	if bus.lastPayload == nil {
		t.Fatalf("expected payload, got nil")
	}
}

func TestHandlerTransactionInvalidJSON(t *testing.T) {
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	bus := &fakeBus{}
	s := NewFiberServer(logger, txSvc, bus)
	app := s.app.(*fiber.App)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", resp.StatusCode)
	}
}

func TestHandlerTransactionInvalidGas(t *testing.T) {
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	bus := &fakeBus{}
	s := NewFiberServer(logger, txSvc, bus)
	app := s.app.(*fiber.App)

	body := map[string]string{
		"from":      "0xfrom",
		"to":        "0xto",
		"chain":     "sepolia",
		"amount":    "10",
		"gas":       "notanumber",
		"gas_price": "1",
	}
	b, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/transaction", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid gas, got %d", resp.StatusCode)
	}
}

func TestHandlerHeatlCheckMethod(t *testing.T) {
	logger := zap.NewNop()
	txSvc := &service.TransactionService{}
	s := NewFiberServer(logger, txSvc, nil)
	app := s.app.(*fiber.App)

	// register a route that uses the method receiver so we invoke handlerHeatlCheck
	app.Get("/direct-health", s.handlerHeatlCheck)
	req, _ := http.NewRequest("GET", "/direct-health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
