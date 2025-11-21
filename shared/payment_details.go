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
	PisNumber    pgtype.Int4
}

type ReversalDetails struct {
	PaymentType      TransactionType
	ErroredCourtRef  pgtype.Text
	CorrectCourtRef  pgtype.Text
	BankDate         pgtype.Date
	ReceivedDate     pgtype.Timestamp
	Amount           int32
	CreatedBy        pgtype.Int4
	PisNumber        pgtype.Int4
	SkipBankDate     pgtype.Bool
	Notes            pgtype.Text
	SkipReceivedDate pgtype.Bool
}

type FulfilledRefundDetails struct {
	CourtRef      pgtype.Text
	Amount        pgtype.Int4
	AccountName   pgtype.Text
	AccountNumber pgtype.Text
	SortCode      pgtype.Text
	UploadedBy    pgtype.Int4
	BankDate      pgtype.Date
}
