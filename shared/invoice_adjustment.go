package shared

type InvoiceAdjustments []InvoiceAdjustment

type InvoiceAdjustment struct {
	Id             int    `json:"id"`
	InvoiceRef     string `json:"invoiceRef"`
	RaisedDate     Date   `json:"raisedDate"`
	AdjustmentType string `json:"adjustmentType"`
	Amount         int    `json:"amount"`
	Status         string `json:"status"`
	Notes          string `json:"notes"`
}
