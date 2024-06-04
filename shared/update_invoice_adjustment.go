package shared

type UpdateInvoiceAdjustment struct {
	Status string `json:"status" validate:"oneof=APPROVED REJECTED"`
}
