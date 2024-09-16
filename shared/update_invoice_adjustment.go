package shared

type UpdateInvoiceAdjustment struct {
	Status AdjustmentStatus `json:"status" validate:"valid-enum,oneof=2 3"` // APPROVED, REJECTED
}
