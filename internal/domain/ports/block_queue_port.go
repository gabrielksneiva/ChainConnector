package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
)

type BlockProducerPort interface {
	Enabled() bool
	EnqueueBlockEvent(ctx context.Context, event *entity.BlockEvent) error
}

type BlockEventProcessorPort interface {
	ProcessBlockEvent(ctx context.Context, event *entity.BlockEvent) error
}
