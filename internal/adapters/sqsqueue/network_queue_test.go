package sqsqueue

import (
	"ChainConnector/internal/config"
	"testing"

	"go.uber.org/zap"
)

func TestNewNetworkQueueDisabled(t *testing.T) {
	queue, err := NewNetworkQueue(&config.Config{}, zap.NewNop())
	if err != nil {
		t.Fatalf("expected disabled queue without error, got %v", err)
	}
	if queue == nil {
		t.Fatal("expected queue")
	}
	if queue.Enabled() {
		t.Fatal("expected producer to be disabled")
	}
}

func TestNewNetworkQueueRequiresQueueWhenEnabled(t *testing.T) {
	_, err := NewNetworkQueue(&config.Config{
		SQSEnabled: true,
		AWSRegion:  "us-east-1",
	}, zap.NewNop())
	if err == nil {
		t.Fatal("expected missing queue configuration error")
	}
}

func TestNewNetworkQueueRejectsPartialStaticCredentials(t *testing.T) {
	_, err := NewNetworkQueue(&config.Config{
		SQSEnabled:       true,
		AWSRegion:        "us-east-1",
		NetworkQueueName: "chainconnector-network-registrations",
		AWSAccessKeyID:   "test",
	}, zap.NewNop())
	if err == nil {
		t.Fatal("expected partial credential configuration error")
	}
}
