package shared

type AddFeeReduction struct {
	FeeType       FeeReductionType `json:"feeType" validate:"required,valid-enum"`
	StartYear     string           `json:"startYear" validate:"required"`
	LengthOfAward int              `json:"lengthOfAward" validate:"required,gte=1,lte=3"`
	DateReceived  *Date            `json:"dateReceived,omitempty" validate:"required,date-in-the-past"`
	Notes         string           `json:"notes" validate:"required,thousand-character-limit"`
}
