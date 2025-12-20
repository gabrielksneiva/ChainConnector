package ports

import (
	"ChainConnector/internal/domain/entity"
	"context"
)

// WalletSignerPort abstracts signing operations. Domain asks for signature, not key management.
type WalletSignerPort interface {
	// SignTransaction signs the domain Transaction and returns raw signed bytes (RLP).
	SignTransaction(ctx context.Context, tx *entity.Transaction) ([]byte, error)

	// Address returns the address controlled by this signer (for tx.From or metadata).
	Address(ctx context.Context) (string, error)

	// Optional: SignHash signs arbitrary hash (useful for EIP-712, auth, etc).
	SignHash(ctx context.Context, hash []byte) ([]byte, error)
}
