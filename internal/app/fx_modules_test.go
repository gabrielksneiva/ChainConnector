package app

import (
	"ChainConnector/internal/config"
	"testing"

	"go.uber.org/zap"
)

func TestProviderETHRPC(t *testing.T) {
	logger := zap.NewNop()
	cfg := config.Load()
	eth := providerETHRPC(logger, cfg)
	if eth == nil {
		t.Fatalf("expected non-nil provider result")
	}
}

func TestNewZapLogger(t *testing.T) {
	l, err := newZapLogger()
	if err != nil {
		t.Fatalf("unexpected error from newZapLogger: %v", err)
	}
	if l == nil {
		t.Fatalf("expected non-nil logger")
	}
}
