package entity

import "math/big"

type ReceiptStatus int

const (
	ReceiptStatusUnknown ReceiptStatus = iota
	ReceiptStatusFailed
	ReceiptStatusSuccess
)

type Receipt struct {
	TxHash            string        `json:"tx_hash" db:"tx_hash"`
	BlockNumber       uint64        `json:"block_number" db:"block_number"`
	BlockHash         string        `json:"block_hash" db:"block_hash"`
	Status            ReceiptStatus `json:"status" db:"status"`
	ContractAddress   string        `json:"contract_address,omitempty" db:"contract_address"`
	Logs              []Log         `json:"logs" db:"logs"`
	GasUsed           uint64        `json:"gas_used" db:"gas_used"`
	CumulativeGasUsed uint64        `json:"cumulative_gas_used" db:"cumulative_gas_used"`
	Root              string        `json:"root,omitempty" db:"root"`
	EffectiveGasPrice *big.Int      `json:"effective_gas_price,omitempty" db:"effective_gas_price"`
}
