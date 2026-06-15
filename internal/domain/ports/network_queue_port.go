package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
)

type NetworkProducerPort interface {
	Enabled() bool
	EnqueueNetworkRegistration(ctx context.Context, network *entity.Network) error
}
