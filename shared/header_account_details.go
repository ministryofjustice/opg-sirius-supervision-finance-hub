package shared

type HeaderAccountData struct {
	ClientID           int    `json:"clientId"`
	OutstandingBalance string `json:"cachedOutstandingBalance"`
	CreditBalance      string `json:"cachedCreditBalance"`
	PaymentMethod      string `json:"paymentMethod"`
}
