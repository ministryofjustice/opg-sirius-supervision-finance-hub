package shared

type AddManualInvoice struct {
	InvoiceType      InvoiceType `json:"invoiceType" validate:"required,valid-enum"`
	Amount           int         `json:"amount" validate:"required,gt=0,lte=32000"`
	RaisedDate       *Date       `json:"raisedDate,omitempty" validate:"required"`
	StartDate        *Date       `json:"StartDate,omitempty" validate:"required"`
	EndDate          *Date       `json:"endDate,omitempty" validate:"required"`
	SupervisionLevel string      `json:"supervisionLevel"`
}
