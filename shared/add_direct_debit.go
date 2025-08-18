package shared

type SetupDirectDebit struct {
	AccountName   string `json:"name" validate:"required"`
	SortCode      string `json:"sortCode" validate:"required,len=8"`
	AccountNumber string `json:"account" validate:"required,numeric,len=8"`
}
