package app

import (
	"testing"

	"go.uber.org/zap"
)

func TestProviderETHRPC(t *testing.T) {
	logger := zap.NewNop()
	eth := providerETHRPC(logger)
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
