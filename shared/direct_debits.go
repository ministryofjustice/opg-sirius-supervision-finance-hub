package shared

type AllPayCustomer struct {
	Surname         string `json:"Surname" validate:"required"`
	ClientReference string `json:"ClientReference" validate:"required"`
}

type AllPayBankDetails struct {
	AccountName   string `json:"accountName" validate:"required,lte=18"`
	SortCode      string `json:"sortCode" validate:"required,len=8"`
	AccountNumber string `json:"accountNumber" validate:"required,len=8,numeric"`
}

type Address struct {
	Line1    string `json:"line1" validate:"required,lte=40"`
	Town     string `json:"town" validate:"required,lte=40"`
	PostCode string `json:"postCode" validate:"required,lte=10"`
}

type CreateMandate struct {
	AllPayCustomer
	Address     Address `json:"address"`
	BankAccount struct {
		BankDetails AllPayBankDetails `json:"bankDetails"`
	} `json:"bankAccount"`
}

type CancelMandate struct {
	AllPayCustomer
}

type CreateSchedule struct {
	AllPayCustomer
}

type CreateScheduleForInvoice struct {
	CreateSchedule
	InvoiceId int32 `json:"invoiceId"`
}
