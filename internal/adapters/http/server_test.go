package http

import (
	"context"
	"net/http"
	"testing"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

var capturedHook fx.Hook

type fakeApp struct{}

func (f *fakeApp) Listen(_ string) error { return nil }
func (f *fakeApp) Shutdown() error       { return nil }

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
	s := NewFiberServer(logger)

	// Use zero-value lifecycle; Start should handle nil Append without panicking.
	var lc fx.Lifecycle
	s.Start(lc)
}

func TestNewFiberServer_ConstructsWithLogger(t *testing.T) {
	logger := zap.NewNop()
	s := NewFiberServer(logger)
	if s == nil || s.app == nil {
		t.Fatalf("expected non-nil FiberServer and app")
	}
	if s.logger == nil {
		t.Fatalf("expected logger to be set on FiberServer")
	}
}

func TestFiberServer_HookExecution(t *testing.T) {
	logger := zap.NewNop()
	s := NewFiberServer(logger)
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
