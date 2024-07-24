package shared

type AddManualInvoice struct {
	InvoiceType      InvoiceType `json:"invoiceType" validate:"required,valid-enum"`
	Amount           int         `json:"amount" validate:"required,gt=0,lte=32000"`
	RaisedDate       *Date       `json:"raisedDate" validate:"omitempty,required"`
	StartDate        *Date       `json:"StartDate" validate:"omitempty,required"`
	EndDate          *Date       `json:"endDate" validate:"omitempty,required"`
	SupervisionLevel string      `json:"supervisionLevel" validate:"oneof=GENERAL MINIMAL"`
}
