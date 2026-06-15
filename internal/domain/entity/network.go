package entity

import "time"

type Network struct {
	ID             string    `json:"id" db:"id"`
	Name           string    `json:"name" db:"name"`
	ChainID        int64     `json:"chain_id" db:"chain_id"`
	RPCURL         string    `json:"rpc_url,omitempty" db:"rpc_url"`
	CurrencySymbol string    `json:"currency_symbol,omitempty" db:"currency_symbol"`
	ExplorerURL    string    `json:"explorer_url,omitempty" db:"explorer_url"`
	Enabled        bool      `json:"enabled" db:"enabled"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
