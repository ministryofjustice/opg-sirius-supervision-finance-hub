package shared

type AddManualInvoice struct {
	InvoiceType      InvoiceType  `json:"invoiceType" validate:"required,valid-enum"`
	Amount           NillableInt  `json:"amount" validate:"int-required-if-not-nil,nillable-int-gt=50,nillable-int-lte=32000"`
	RaisedDate       NillableDate `json:"raisedDate" validate:"date-required-if-not-nil"`
	RaisedYear       NillableInt  `json:"raisedYear" validate:"int-required-if-not-nil"`
	StartDate        NillableDate `json:"startDate" validate:"date-required-if-not-nil"`
	EndDate          NillableDate `json:"endDate" validate:"date-required-if-not-nil"`
	SupervisionLevel string       `json:"supervisionLevel" validate:"omitempty,oneof=GENERAL MINIMAL"`
}
