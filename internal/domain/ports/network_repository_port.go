package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
)

type NetworkRepositoryPort interface {
	SaveNetwork(ctx context.Context, network *entity.Network) error
	FindNetworkByID(ctx context.Context, id string) (*entity.Network, error)
	ListNetworks(ctx context.Context) ([]*entity.Network, error)
}
