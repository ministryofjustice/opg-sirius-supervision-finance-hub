package shared

type AddManualInvoice struct {
	InvoiceType      InvoiceType      `json:"invoiceType" validate:"required,valid-enum"`
	Amount           Nillable[int32]  `json:"amount" validate:"nillable-int-gt=50,nillable-int-lte=32000"`
	RaisedDate       Nillable[Date]   `json:"raisedDate" validate:"nillable-date-required"`
	StartDate        Nillable[Date]   `json:"startDate" validate:"nillable-date-required"`
	EndDate          Nillable[Date]   `json:"endDate" validate:"nillable-date-required"`
	SupervisionLevel Nillable[string] `json:"supervisionLevel" validate:"nillable-string-oneof=GENERAL MINIMAL"`
}
