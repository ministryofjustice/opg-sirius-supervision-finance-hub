package shared

type AnnualBillingInformation struct {
	AnnualBillingYear        string
	DemandedExpectedCount    int
	DemandedIssuedCount      int
	DemandedSkippedCount     int
	DirectDebitExpectedCount int
	DirectDebitIssuedCount   int
}
