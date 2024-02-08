package model

type FinancePerson struct {
	ID                 int    `json:"id"`
	Firstname          string `json:"firstname"`
	Surname            string `json:"surname"`
	CourtRef           string `json:"caseRecNumber"`
	OutstandingBalance string `json:"outstandingBalance"`
	CreditBalance      string `json:"creditBalance"`
	PaymentMethod      string `json:"paymentMethod"`
}
