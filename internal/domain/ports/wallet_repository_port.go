package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
)

type WalletRepositoryPort interface {
	SaveWallet(ctx context.Context, wallet *entity.Wallet) error
	FindWalletByID(ctx context.Context, id string) (*entity.Wallet, error)
	FindWalletByAddress(ctx context.Context, address string) (*entity.Wallet, error)
	ListWallets(ctx context.Context, chain string) ([]*entity.Wallet, error)
	ListAllWallets(ctx context.Context) ([]*entity.Wallet, error)
}
