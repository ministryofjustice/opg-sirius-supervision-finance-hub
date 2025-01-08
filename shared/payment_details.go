package shared

import "time"

type PaymentDetails struct {
	Amount     int32
	BankDate   time.Time
	CourtRef   string
	LedgerType string
	UploadDate time.Time
}
