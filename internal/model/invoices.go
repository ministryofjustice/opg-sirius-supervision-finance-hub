package model

type InvoiceList struct {
	Invoices []Invoice `json:"invoices"`
}

type Invoice struct {
	Id                 int                `json:"id"`
	Ref                string             `json:"ref"`
	Status             string             `json:"status"`
	Amount             string             `json:"amount"`
	RaisedDate         string             `json:"raisedDate"`
	Received           string             `json:"received"`
	OutstandingBalance string             `json:"outstandingBalance"`
	Ledgers            []Ledger           `json:"ledgers"`
	SupervisionLevels  []SupervisionLevel `json:"supervisionLevels"`
}

type Ledger struct {
	Amount          string `json:"amount"`
	ReceivedDate    string `json:"receivedDate"`
	TransactionType string `json:"transactionType"`
	Status          string `json:"status"`
}

type SupervisionLevel struct {
	Level  string `json:"level"`
	Amount string `json:"amount"`
	From   string `json:"from"`
	To     string `json:"to"`
}

type T2 struct {
	Invoice []struct {
		Id                 int    `json:"id"`
		Ref                string `json:"ref"`
		Status             string `json:"status"`
		Amount             string `json:"amount"`
		RaisedDate         string `json:"raisedDate"`
		Received           string `json:"received"`
		OutstandingBalance string `json:"outstandingBalance"`
	} `json:"invoice"`
}

type T3 struct {
	Invoice struct {
		Id                 int    `json:"id"`
		Ref                string `json:"ref"`
		Status             string `json:"status"`
		Amount             string `json:"amount"`
		RaisedDate         string `json:"raisedDate"`
		Received           string `json:"received"`
		OutstandingBalance string `json:"outstandingBalance"`
	} `json:"invoice"`
}

type T4 struct {
	Invoices []struct {
		Invoice struct {
			Id                 int    `json:"id"`
			Ref                string `json:"ref"`
			Status             string `json:"status"`
			Amount             string `json:"amount"`
			RaisedDate         string `json:"raisedDate"`
			Received           string `json:"received"`
			OutstandingBalance string `json:"outstandingBalance"`
		} `json:"invoice"`
	} `json:"invoices"`
}
