package shared

type AllPayBankDetails struct {
	AccountName   string `json:"AccountName"`
	SortCode      string `json:"SortCode"`
	AccountNumber string `json:"AccountNumber"`
}

type Address struct {
	Line1    string `json:"Line1"`
	Town     string `json:"Town"`
	PostCode string `json:"PostCode"`
}

type CreateMandate struct {
	SchemeCode      string  `json:"SchemeCode"`
	ClientReference string  `json:"ClientReference"`
	Surname         string  `json:"LastName"`
	Address         Address `json:"Address"`
	BankAccount     struct {
		BankDetails AllPayBankDetails `json:"BankDetails"`
	} `json:"BankAccount"`
}
