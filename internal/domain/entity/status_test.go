package entity

import "testing"

// func (s TxStatus) String() string {
// 	switch s {
// 	case TxStatusPending:
// 		return "pending"
// 	case TxStatusSigned:
// 		return "signed"
// 	case TxStatusSent:
// 		return "sent"
// 	case TxStatusConfirmed:
// 		return "confirmed"
// 	case TxStatusFailed:
// 		return "failed"
// 	case TxStatusCancelled:
// 		return "cancelled"
// 	default:
// 		return "unknown"
// 	}
// }

func TestTxStatusString(t *testing.T) {
	tests := []struct {
		status   TxStatus
		expected string
	}{
		{TxStatusPending, "pending"},
		{TxStatusSigned, "signed"},
		{TxStatusSent, "sent"},
		{TxStatusConfirmed, "confirmed"},
		{TxStatusFailed, "failed"},
		{TxStatusCancelled, "cancelled"},
		{TxStatusUnknown, "unknown"},
		{TxStatus(999), "unknown"}, // Test for an undefined status
	}

	for _, test := range tests {
		result := test.status.String()
		if result != test.expected {
			t.Errorf("TxStatus.String() for status %d: expected %s, got %s", test.status, test.expected, result)
		}
	}
}
