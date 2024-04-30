package shared

import (
	"encoding/json"
	"fmt"
	"strings"
)

var InvoiceTypes = []InvoiceType{
	WriteOff,
	AddCredit,
	AddDebit,
	Unapply,
	Reapply,
}

type InvoiceType int

const (
	Unknown InvoiceType = iota
	WriteOff
	AddCredit
	AddDebit
	Unapply
	Reapply
)

var invoiceTypeMap = map[string]InvoiceType{
	"CREDIT WRITE OFF": WriteOff,
	"CREDIT MEMO":      AddCredit,
	"UNKNOWN DEBIT":    AddDebit,
	"UNAPPLY":          Unapply,
	"REAPPLY":          Reapply,
}

func (i InvoiceType) Translation() string {
	switch i {
	case WriteOff:
		return "Write off"
	case AddCredit:
		return "Add credit"
	case AddDebit:
		return "Add debit"
	case Unapply:
		return "Unapply"
	case Reapply:
		return "Reapply"
	default:
		return ""
	}
}

func (i InvoiceType) Key() string {
	switch i {
	case WriteOff:
		return "CREDIT WRITE OFF"
	case AddCredit:
		return "CREDIT MEMO"
	case AddDebit:
		return "UNKNOWN DEBIT"
	case Unapply:
		return "UNAPPLY"
	case Reapply:
		return "REAPPLY"
	default:
		return ""
	}
}

func (i InvoiceType) AmountRequired() bool {
	switch i {
	case AddCredit, AddDebit, Unapply, Reapply:
		return true
	default:
		return false
	}
}

func (i InvoiceType) IsValid() bool {
	switch i {
	case AddCredit, AddDebit, Unapply, Reapply:
		return true
	default:
		return false
	}
}

func parseInvoiceType(s string) (InvoiceType, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	value, ok := invoiceTypeMap[s]
	if !ok {
		return InvoiceType(0), fmt.Errorf("%q is not a valid card suit", s)
	}
	return value, nil
}

func (i InvoiceType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *InvoiceType) UnmarshalJSON(data []byte) (err error) {
	var suits string
	if err := json.Unmarshal(data, &suits); err != nil {
		return err
	}
	if *i, err = parseInvoiceType(suits); err != nil {
		return err
	}
	return nil
}
