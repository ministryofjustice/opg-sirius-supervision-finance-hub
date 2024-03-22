package shared

type AccountInformation struct {
	OutstandingBalance int    `json:"outstandingBalance"`
	CreditBalance      int    `json:"creditBalance"`
	PaymentMethod      string `json:"paymentMethod"`
}
