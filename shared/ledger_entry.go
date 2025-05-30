package shared

type AddInvoiceAdjustmentRequest struct {
	AdjustmentType  AdjustmentType `json:"adjustmentType" validate:"valid-enum"`
	AdjustmentNotes string         `json:"notes" validate:"required,thousand-character-limit"`
	Amount          int32          `json:"amount,omitempty" validate:"required_if=AdjustmentType 2,required_if=AdjustmentType 3,omitempty,gt=0"`
	ManagerOverride bool           `json:"managerOverride,omitempty"`
}
