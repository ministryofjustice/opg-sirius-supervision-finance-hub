package shared

type AddManualInvoice struct {
	InvoiceType      InvoiceType    `json:"invoiceType" validate:"required,valid-enum"`
	Amount           NillableInt    `json:"amount" validate:"nillable-int-required,nillable-int-gt=50,nillable-int-lte=32000"`
	RaisedDate       NillableDate   `json:"raisedDate" validate:"nillable-date-required"`
	RaisedYear       NillableInt    `json:"raisedYear" validate:"nillable-int-required"`
	StartDate        NillableDate   `json:"startDate" validate:"nillable-date-required"`
	EndDate          NillableDate   `json:"endDate" validate:"nillable-date-required"`
	SupervisionLevel NillableString `json:"supervisionLevel" validate:"nillable-string-oneof=GENERAL MINIMAL"`
}
