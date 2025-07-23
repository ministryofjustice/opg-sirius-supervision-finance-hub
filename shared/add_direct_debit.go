package shared

type AddDirectDebit struct {
	AccountHolder string `json:"accountHolder" validate:"required"`
	AccountName   string `json:"name" validate:"required"`
	SortCode      string `json:"sortCode" validate:"required,len=8"`
	AccountNumber string `json:"account" validate:"required,numeric,len=8"`
}
