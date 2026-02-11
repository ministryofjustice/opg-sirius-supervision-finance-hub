package shared

type AnnualBillingInformation struct {
	AnnualBillingYear string
	ExpectedCount     int64
	IssuedCount       int64
	SkippedCount      int64
}
