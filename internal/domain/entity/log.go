package entity

type Log struct {
	Address     string   `json:"address" db:"address"`
	Topics      []string `json:"topics" db:"topics"`
	Data        []byte   `json:"data" db:"data"`
	BlockNumber uint64   `json:"block_number" db:"block_number"`
	TxHash      string   `json:"tx_hash" db:"tx_hash"`
	LogIndex    uint32   `json:"log_index" db:"log_index"`
}

type LogFilter struct {
	FromBlock *uint64    `json:"from_block,omitempty"`
	ToBlock   *uint64    `json:"to_block,omitempty"`
	Addresses []string   `json:"addresses,omitempty"`
	Topics    [][]string `json:"topics,omitempty"` // outer slice = AND, inner slice = OR
}
