package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"math/big"
)

type TxRepositoryPort interface {
	Save(ctx context.Context, tx *entity.Transaction) error
	FindByID(ctx context.Context, id string) (*entity.Transaction, error)
	FindByHash(ctx context.Context, hash string) (*entity.Transaction, error)
	UpdateStatus(ctx context.Context, txID string, status entity.TxStatus, updates map[string]interface{}) error
	ListPending(ctx context.Context, limit int) ([]*entity.Transaction, error)

	// Balance management
	GetBalance(ctx context.Context, address string, chain string) (*big.Int, error)
	UpdateBalance(ctx context.Context, address string, chain string, amount *big.Int) error

	// Interest addresses management
	AddInterestAddress(ctx context.Context, address string, chain string) error
	GetInterestAddresses(ctx context.Context, chain string) ([]string, error)
}
