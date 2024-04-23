package shared

type AddFeeReduction struct {
	ClientId          int    `json:"clientId"`
	FeeType           string `json:"feeType"`
	StartYear         string `json:"startYear"`
	LengthOfAward     string `json:"lengthOfAward"`
	DateReceive       string `json:"dateReceived"`
	FeeReductionNotes string `json:"feeReductionNotes"`
}
