package entity

import (
	"math/big"
	"time"
)

type Transaction struct {
	ID   string `json:"id" db:"id"`
	From string `json:"from" db:"from"`
	// Chain is the logical chain name (e.g. "ETH", "POLYGON") used to select RPC.
	Chain    string   `json:"chain,omitempty" db:"chain"`
	To       *string  `json:"to,omitempty" db:"to"`
	Value    *big.Int `json:"value" db:"value"`
	Gas      uint64   `json:"gas" db:"gas"`
	GasPrice *big.Int `json:"gas_price" db:"gas_price"`
	// EIP-1559 fields (optional). If set, signer should produce a DynamicFeeTx.
	MaxPriorityFeePerGas *big.Int  `json:"max_priority_fee_per_gas,omitempty" db:"max_priority_fee_per_gas"`
	MaxFeePerGas         *big.Int  `json:"max_fee_per_gas,omitempty" db:"max_fee_per_gas"`
	Nonce                uint64    `json:"nonce" db:"nonce"`
	Data                 []byte    `json:"data,omitempty" db:"data"`
	ChainID              *big.Int  `json:"chain_id" db:"chain_id"`
	RawTxHex             string    `json:"raw_tx_hex,omitempty" db:"raw_tx_hex"`
	TxHash               string    `json:"tx_hash,omitempty" db:"tx_hash"`
	Status               TxStatus  `json:"status" db:"status"`
	CreatedAt            time.Time `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time `json:"updated_at" db:"updated_at"`

	// Optional lifecycle timestamps
	SentAt      *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	ConfirmedAt *time.Time `json:"confirmed_at,omitempty" db:"confirmed_at"`

	// Failure reason
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`
}
