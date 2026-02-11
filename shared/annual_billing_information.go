package shared

import "github.com/jackc/pgx/v5/pgtype"

type AnnualBillingInformation struct {
	AnnualBillingYear string
	ExpectedCount     pgtype.Int8
	IssuedCount       pgtype.Int8
	SkippedCount      pgtype.Int8
}
