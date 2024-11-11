package shared

import "encoding/json"

type TransactionType int

const (
	TransactionTypeUnknown TransactionType = iota
	TransactionTypeWriteOff
	TransactionTypeCreditMemo
	TransactionTypeDebitMemo
	TransactionTypeWriteOffReversal
	TransactionTypeExemption
	TransactionTypeHardship
	TransactionTypeRemission
	TransactionTypeReapply
	TransactionTypeMotoCardPayment
	TransactionTypeOnlineCardPayment
	TransactionTypeSupervisionBACSPayment
)

var TransactionTypeMap = map[string]TransactionType{
	"CREDIT WRITE OFF":         TransactionTypeWriteOff,
	"CREDIT MEMO":              TransactionTypeCreditMemo,
	"DEBIT MEMO":               TransactionTypeDebitMemo,
	"WRITE OFF REVERSAL":       TransactionTypeWriteOffReversal,
	"EXEMPTION":                TransactionTypeExemption,
	"HARDSHIP":                 TransactionTypeHardship,
	"REMISSION":                TransactionTypeRemission,
	"CREDIT REAPPLY":           TransactionTypeReapply,
	"MOTO CARD PAYMENT":        TransactionTypeMotoCardPayment,
	"ONLINE CARD PAYMENT":      TransactionTypeOnlineCardPayment,
	"SUPERVISION BACS PAYMENT": TransactionTypeSupervisionBACSPayment,
}

func (t TransactionType) String() string {
	switch t {
	case TransactionTypeWriteOff:
		return "Write off"
	case TransactionTypeCreditMemo:
		return "Credit memo"
	case TransactionTypeDebitMemo:
		return "Debit memo"
	case TransactionTypeWriteOffReversal:
		return "Write off reversal"
	case TransactionTypeExemption:
		return "Exemption"
	case TransactionTypeHardship:
		return "Hardship"
	case TransactionTypeRemission:
		return "Remission"
	case TransactionTypeReapply:
		return "Reapply"
	case TransactionTypeMotoCardPayment:
		return "MOTO card payment"
	case TransactionTypeOnlineCardPayment:
		return "Online card payment"
	case TransactionTypeSupervisionBACSPayment:
		return "BACS payment (Supervision account)"
	default:
		return ""
	}
}

func (t TransactionType) Key() string {
	switch t {
	case TransactionTypeWriteOff:
		return "CREDIT WRITE OFF"
	case TransactionTypeCreditMemo:
		return "CREDIT MEMO"
	case TransactionTypeDebitMemo:
		return "DEBIT MEMO"
	case TransactionTypeWriteOffReversal:
		return "WRITE OFF REVERSAL"
	case TransactionTypeExemption:
		return "EXEMPTION"
	case TransactionTypeHardship:
		return "HARDSHIP"
	case TransactionTypeRemission:
		return "REMISSION"
	case TransactionTypeReapply:
		return "REAPPLY"
	case TransactionTypeMotoCardPayment:
		return "MOTO CARD PAYMENT"
	case TransactionTypeOnlineCardPayment:
		return "ONLINE CARD PAYMENT"
	case TransactionTypeSupervisionBACSPayment:
		return "SUPERVISION BACS PAYMENT"
	default:
		return ""
	}
}

func (t TransactionType) Valid() bool {
	return t != TransactionTypeUnknown
}

func ParseTransactionType(s string) TransactionType {
	value, ok := TransactionTypeMap[s]
	if !ok {
		return TransactionType(0)
	}
	return value
}

func (t TransactionType) MarshalJSON() ([]byte, error) {
	return json.Marshal(t.Key())
}

func (t *TransactionType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*t = ParseTransactionType(s)
	return nil
}
