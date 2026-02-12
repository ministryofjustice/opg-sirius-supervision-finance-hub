package shared

type AnnualBillingInformation struct {
	AnnualBillingYear        string
	DemandedExpectedCount    int64
	DemandedIssuedCount      int64
	DemandedSkippedCount     int64
	DirectDebitExpectedCount int64
	DirectDebitIssuedCount   int64
}
