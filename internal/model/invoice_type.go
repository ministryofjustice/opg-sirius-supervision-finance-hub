package model

type InvoiceType struct {
	Handle      string
	Description string
}

var InvoiceTypes = []InvoiceType{
	{Handle: "writeOff", Description: "Write off"},
	{Handle: "addCredit", Description: "Add credit"},
	{Handle: "addDebit", Description: "Add debit"},
	{Handle: "unapply", Description: "Unapply"},
	{Handle: "reapply", Description: "Reapply"},
}
