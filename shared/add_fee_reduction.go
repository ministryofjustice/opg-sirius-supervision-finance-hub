package shared

type AddFeeReduction struct {
	ClientId          int    `json:"clientId"`
	FeeType           string `json:"feeType"`
	StartYear         string `json:"startYear"`
	LengthOfAward     string `json:"lengthOfAward"`
	DateReceive       Date   `json:"dateReceived"`
	FeeReductionNotes string `json:"feeReductionNotes"`
}
