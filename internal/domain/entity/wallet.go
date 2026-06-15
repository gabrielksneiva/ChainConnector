package entity

import "time"

type Wallet struct {
	ID         string    `json:"id" db:"id"`
	Address    string    `json:"address" db:"address"`
	Chain      string    `json:"chain" db:"chain"`
	PrivateKey string    `json:"-" db:"private_key"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}
