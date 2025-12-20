package entity

import (
	"testing"
	"time"
)

func TestEventTypesAndTimestamp(t *testing.T) {
	now := time.Now()
	be := BaseEvent{When: now}
	if be.Timestamp() != now {
		t.Fatalf("expected timestamp to match")
	}

	var e1 TxCreatedEvent
	if e1.Type() != "TxCreated" {
		t.Fatalf("unexpected type %s", e1.Type())
	}
	var e2 TxSignedEvent
	if e2.Type() != "TxSigned" {
		t.Fatalf("unexpected type %s", e2.Type())
	}
	var e3 TxSentEvent
	if e3.Type() != "TxSent" {
		t.Fatalf("unexpected type %s", e3.Type())
	}
	var e4 TxConfirmedEvent
	if e4.Type() != "TxConfirmed" {
		t.Fatalf("unexpected type %s", e4.Type())
	}
	var e5 TxFailedEvent
	if e5.Type() != "TxFailed" {
		t.Fatalf("unexpected type %s", e5.Type())
	}
}
