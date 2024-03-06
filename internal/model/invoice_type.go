package model

type InvoiceType struct {
	Handle         string
	Description    string
	AmountRequired bool
}

var InvoiceTypes = []InvoiceType{
	{Handle: "writeOff", Description: "Write off", AmountRequired: false},
	{Handle: "addCredit", Description: "Add credit", AmountRequired: true},
	{Handle: "addDebit", Description: "Add debit", AmountRequired: true},
	{Handle: "unapply", Description: "Unapply", AmountRequired: true},
	{Handle: "reapply", Description: "Reapply", AmountRequired: true},
}
