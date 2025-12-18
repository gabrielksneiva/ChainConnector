package app

import (
	"testing"
)

func TestNewZapLogger(t *testing.T) {
	l, err := newZapLogger()
	if err != nil {
		t.Fatalf("unexpected error from newZapLogger: %v", err)
	}
	if l == nil {
		t.Fatalf("expected non-nil logger")
	}
}
