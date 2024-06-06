package shared

type BillingHistory struct {
	InvoiceReference   string `json:"invoice_reference"`
	User               string `json:"user"`
	Date               Date   `json:"date"`
	EventType          string `json:"event_type"`
	Type               string `json:"type"` // type of invoice/fee reduction/ledger type
	Amount             int    `json:"amount"`
	OutstandingBalance int    `json:"outstanding_balance"`
}
