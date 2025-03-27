package shared

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type PaymentDetails struct {
	Amount       int32
	BankDate     pgtype.Date
	CourtRef     pgtype.Text
	LedgerType   TransactionType
	ReceivedDate pgtype.Timestamp
	CreatedBy    pgtype.Int4
}
