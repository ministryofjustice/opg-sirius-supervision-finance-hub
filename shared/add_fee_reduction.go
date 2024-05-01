package shared

type AddFeeReduction struct {
	ClientId          int    `json:"clientId"`
	FeeType           string `json:"feeType" validate:"required"`
	StartYear         string `json:"startYear" validate:"required"`
	LengthOfAward     string `json:"lengthOfAward" validate:"required"`
	DateReceive       Date   `json:"dateReceived" validate:"required,check-in-the-past"`
	FeeReductionNotes string `json:"feeReductionNotes" validate:"required,check-character-limit"`
}
