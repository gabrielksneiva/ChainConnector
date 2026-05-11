package entity

import (
	"time"
)

type UserBalance struct {
	ID        string    `json:"id" db:"id"`
	Address   string    `json:"address" db:"address"`
	Chain     string    `json:"chain" db:"chain"`
	Balance   string    `json:"balance" db:"balance"` // Using string for big numbers
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type InterestAddress struct {
	ID        string    `json:"id" db:"id"`
	Address   string    `json:"address" db:"address"`
	Chain     string    `json:"chain" db:"chain"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
