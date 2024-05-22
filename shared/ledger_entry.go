package shared

type CreateLedgerEntryRequest struct {
	AdjustmentType  AdjustmentType `json:"adjustmentType" validate:"valid-enum"`
	AdjustmentNotes string         `json:"notes" validate:"required,thousand-character-limit"`
	Amount          int            `json:"amount,omitempty" validate:"required_if=AdjustmentType 2,required_if=AdjustmentType 3,omitempty,gt=0"`
}
