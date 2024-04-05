package shared

type Invoices []Invoice

type Invoice struct {
	Id                 int                `json:"id"`
	Ref                string             `json:"ref"`
	Status             string             `json:"status"`
	Amount             int                `json:"amount"`
	RaisedDate         Date               `json:"raisedDate"`
	Received           int                `json:"received"`
	OutstandingBalance int                `json:"outstandingBalance"`
	Ledgers            []Ledger           `json:"ledgers"`
	SupervisionLevels  []SupervisionLevel `json:"supervisionLevels"`
}

type Ledger struct {
	Amount          int    `json:"amount"`
	ReceivedDate    Date   `json:"receivedDate"`
	TransactionType string `json:"transactionType"`
	Status          string `json:"status"`
}

type SupervisionLevel struct {
	Level  string `json:"level"`
	Amount string `json:"amount"`
	From   Date   `json:"from"`
	To     Date   `json:"to"`
}
