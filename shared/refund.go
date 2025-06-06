package shared

type Refunds []Refund

type Refund struct {
	ID            int                   `json:"id"`
	RaisedDate    Date                  `json:"raisedDate"`
	FulfilledDate Nillable[Date]        `json:"fulfilledDate"`
	Amount        int                   `json:"amount"`
	Status        string                `json:"status"`
	Notes         string                `json:"notes"`
	BankDetails   Nillable[BankDetails] `json:"bankDetails"`
	CreatedBy     int                   `json:"createdBy"`
}

type BankDetails struct {
	Name     string `json:"name" validate:"required"`
	Account  string `json:"account" validate:"required,numeric,len=8"`
	SortCode string `json:"sortCode" validate:"required,len=8"`
}

type AddRefund struct {
	AccountName string `json:"name" validate:"required"`
	Account     string `json:"account" validate:"required,numeric,len=8"`
	SortCode    string `json:"sortCode" validate:"required,len=8"`
	Notes       string `json:"notes" validate:"required,thousand-character-limit"`
}
