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
	PisNumber    Nillable[int]
}

type ReversalDetails struct {
	PaymentType     TransactionType
	ErroredCourtRef pgtype.Text
	CorrectCourtRef pgtype.Text
	BankDate        pgtype.Date
	ReceivedDate    pgtype.Timestamp
	Amount          int32
	CreatedBy       pgtype.Int4
}
