package entity

type TxStatus int

const (
	TxStatusUnknown TxStatus = iota
	TxStatusPending
	TxStatusSigned
	TxStatusSent
	TxStatusConfirmed
	TxStatusFailed
	TxStatusCancelled
)

func (s TxStatus) String() string {
	switch s {
	case TxStatusPending:
		return "pending"
	case TxStatusSigned:
		return "signed"
	case TxStatusSent:
		return "sent"
	case TxStatusConfirmed:
		return "confirmed"
	case TxStatusFailed:
		return "failed"
	case TxStatusCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}
