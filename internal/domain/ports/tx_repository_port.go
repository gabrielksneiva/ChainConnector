package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
)

type TxRepositoryPort interface {
	Save(ctx context.Context, tx *entity.Transaction) error
	FindByID(ctx context.Context, id string) (*entity.Transaction, error)
	FindByHash(ctx context.Context, hash string) (*entity.Transaction, error)
	UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error
	ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error)
}
