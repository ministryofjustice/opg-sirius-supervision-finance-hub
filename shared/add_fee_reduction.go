package shared

type AddFeeReduction struct {
	FinanceClientId   int    `json:"financeClientId"`
	FeeType           string `json:"feeType"`
	StartYear         string `json:"startYear"`
	LengthOfAward     string `json:"lengthOfAward"`
	DateReceive       string `json:"dateReceived"`
	FeeReductionNotes string `json:"feeReductionNotes"`
}
