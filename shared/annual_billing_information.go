package shared

type AnnualBillingInformation struct {
	AnnualBillingYear        int
	DemandedExpectedCount    int
	DemandedIssuedCount      int
	DemandedSkippedCount     int
	DirectDebitExpectedCount int
	DirectDebitIssuedCount   int
}
