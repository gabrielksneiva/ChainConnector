package entity

import "time"

type Event interface {
	Type() string
	Timestamp() time.Time
}

type BaseEvent struct {
	When time.Time `json:"timestamp"`
}

func (b BaseEvent) Timestamp() time.Time { return b.When }

type TxCreatedEvent struct {
	BaseEvent
	TxID string `json:"tx_id"`
}

func (TxCreatedEvent) Type() string { return "TxCreated" }

type TxSignedEvent struct {
	BaseEvent
	TxID   string `json:"tx_id"`
	TxHash string `json:"tx_hash"`
}

func (TxSignedEvent) Type() string { return "TxSigned" }

type TxSentEvent struct {
	BaseEvent
	TxID   string `json:"tx_id"`
	TxHash string `json:"tx_hash"`
}

func (TxSentEvent) Type() string { return "TxSent" }

type TxConfirmedEvent struct {
	BaseEvent
	TxID    string  `json:"tx_id"`
	TxHash  string  `json:"tx_hash"`
	Receipt Receipt `json:"receipt"`
}

func (TxConfirmedEvent) Type() string { return "TxConfirmed" }

type TxFailedEvent struct {
	BaseEvent
	TxID  string `json:"tx_id"`
	Error string `json:"error"`
}

func (TxFailedEvent) Type() string { return "TxFailed" }
