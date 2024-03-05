package shared

type Invoices []Invoice

type Invoice struct {
	Id                 int                `json:"id"`
	Ref                string             `json:"ref"`
	Status             string             `json:"status"`
	Amount             string             `json:"amount"`
	RaisedDate         Date               `json:"raisedDate"`
	Received           string             `json:"received"`
	OutstandingBalance string             `json:"outstandingBalance"`
	Ledgers            []Ledger           `json:"ledgers"`
	SupervisionLevels  []SupervisionLevel `json:"supervisionLevels"`
}

type Ledger struct {
	Amount          string `json:"amount"`
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
