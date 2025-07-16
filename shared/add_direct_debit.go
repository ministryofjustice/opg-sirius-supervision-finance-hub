package shared

type AddDirectDebit struct {
	AccountHolder   string           `json:"accountHolder" validate:"required"`
	AccountName     string           `json:"accountName" validate:"required"`
	SortCode        string           `json:"sortCode" validate:"required"`
	AccountNumber   string           `json:"accountNumber" validate:"required"`
}
