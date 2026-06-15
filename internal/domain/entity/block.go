package entity

import (
	"math/big"
	"time"
)

type BlockEvent struct {
	Chain       string    `json:"chain"`
	BlockNumber uint64    `json:"block_number"`
	BlockHash   string    `json:"block_hash,omitempty"`
	ReceivedAt  time.Time `json:"received_at"`
}

type Block struct {
	Chain        string             `json:"chain"`
	Number       uint64             `json:"number"`
	Hash         string             `json:"hash"`
	ParentHash   string             `json:"parent_hash"`
	Transactions []BlockTransaction `json:"transactions"`
}

type BlockTransaction struct {
	Hash     string   `json:"hash"`
	From     string   `json:"from"`
	To       string   `json:"to,omitempty"`
	Value    *big.Int `json:"value,omitempty"`
	Gas      uint64   `json:"gas,omitempty"`
	GasPrice *big.Int `json:"gas_price,omitempty"`
	Nonce    uint64   `json:"nonce,omitempty"`
}
