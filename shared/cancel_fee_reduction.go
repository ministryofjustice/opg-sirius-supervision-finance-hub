package shared

type CancelFeeReduction struct {
	CancellationReason string `json:"cancellationReason" validate:"required,thousand-character-limit"`
}
