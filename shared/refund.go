package shared

type Refunds []Refund

type Refund struct {
	ID            int         `json:"id"`
	RaisedDate    Date        `json:"raisedDate"`
	FulfilledDate Date        `json:"fulfilledDate"`
	Amount        int         `json:"amount"`
	Status        string      `json:"status"`
	Notes         string      `json:"notes"`
	BankDetails   BankDetails `json:"bankDetails"`
}

type BankDetails struct {
	Name     string `json:"name"`
	Account  string `json:"account"`
	SortCode string `json:"sortCode"`
}
