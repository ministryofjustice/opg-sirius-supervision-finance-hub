package shared

type CancelFeeReduction struct {
	Notes string `json:"notes" validate:"required,thousand-character-limit"`
}
