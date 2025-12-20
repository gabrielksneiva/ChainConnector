package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
	"math/big"
)

// BlockchainPort provides read-only blockchain operations used by domain.
type BlockchainPort interface {
	// GetBalance returns the native balance for an address.
	GetBalance(ctx context.Context, address string) (*big.Int, error)

	// GetNonce returns the pending nonce for an address (for tx creation).
	GetNonce(ctx context.Context, address string) (uint64, error)

	// GetTransactionReceipt returns the receipt for a txHash if available.
	GetTransactionReceipt(ctx context.Context, txHash string) (*entity.Receipt, error)

	// GetLogs returns logs matching the provided filter (blocks, topics, address).
	GetLogs(ctx context.Context, f entity.LogFilter) ([]entity.Log, error)

	// GetBlockNumber returns the latest block number.
	GetBlockNumber(ctx context.Context) (uint64, error)

	// EstimateFees returns an estimated priority fee (tip) and max fee (fee cap) in wei for
	// EIP-1559 transactions. The `chain` parameter may be used by implementations that route
	// multiple chains. If unsupported, return an error (e.g., ErrUnsupported).
	EstimateFees(ctx context.Context, chain string) (*big.Int, *big.Int, error)

	// SendRawTransaction sends a fully-signed transaction bytes to the node for the given chain.
	// Returns the transaction hash (hex, with 0x) or an error.
	SendRawTransaction(ctx context.Context, chain string, signedTx []byte) (txHash string, err error)

	// Optional: convenience when you already have hex-encoded signed tx.
	SendRawTransactionHex(ctx context.Context, chain string, signedTxHex string) (txHash string, err error)
}
