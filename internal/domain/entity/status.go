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
func ParseTxStatus(value string) TxStatus {
	switch value {
	case "pending":
		return TxStatusPending
	case "signed":
		return TxStatusSigned
	case "sent":
		return TxStatusSent
	case "confirmed":
		return TxStatusConfirmed
	case "failed":
		return TxStatusFailed
	case "cancelled":
		return TxStatusCancelled
	default:
		return TxStatusUnknown
	}
}
