package shared

type NillableInt struct {
	Value int
	Valid bool
}

type NillableDate struct {
	Value Date
	Valid bool
}

type AddManualInvoice struct {
	InvoiceType      InvoiceType  `json:"invoiceType" validate:"required,valid-enum"`
	Amount           NillableInt  `json:"amount" validate:"int-required-if-not-nil"`
	RaisedDate       NillableDate `json:"raisedDate" validate:"date-required-if-not-nil"`
	StartDate        NillableDate `json:"startDate" validate:"date-required-if-not-nil"`
	EndDate          NillableDate `json:"endDate" validate:"date-required-if-not-nil"`
	SupervisionLevel string       `json:"supervisionLevel" validate:"omitempty,oneof=GENERAL MINIMAL"`
}
