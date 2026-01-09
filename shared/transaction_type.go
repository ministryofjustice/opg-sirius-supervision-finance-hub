package shared

import (
	"encoding/json"
	"slices"
)

type TransactionType int

var paymentTransactionTypes = []TransactionType{
	TransactionTypeMotoCardPayment,
	TransactionTypeOnlineCardPayment,
	TransactionTypeSupervisionBACSPayment,
	TransactionTypeOPGBACSPayment,
	TransactionTypeDirectDebitPayment,
	TransactionTypeSupervisionChequePayment,
	TransactionTypeSOPUnallocatedPayment,
	TransactionTypeRefund,
}

var reversalTransactionTypes = []TransactionType{
	TransactionTypeFeeReductionReversal,
	TransactionTypeWriteOffReversal,
}

const (
	TransactionTypeUnknown TransactionType = iota
	TransactionTypeWriteOff
	TransactionTypeCreditMemo
	TransactionTypeDebitMemo
	TransactionTypeWriteOffReversal
	TransactionTypeFeeReductionReversal
	TransactionTypeExemption
	TransactionTypeHardship
	TransactionTypeRemission
	TransactionTypeReapply
	TransactionTypeMotoCardPayment
	TransactionTypeOnlineCardPayment
	TransactionTypeSupervisionBACSPayment
	TransactionTypeOPGBACSPayment
	TransactionTypeDirectDebitPayment
	TransactionTypeSupervisionChequePayment
	TransactionTypeSOPUnallocatedPayment
	TransactionTypeRefund
)

var TransactionTypeMap = map[string]TransactionType{
	"CREDIT WRITE OFF":           TransactionTypeWriteOff,
	"CREDIT MEMO":                TransactionTypeCreditMemo,
	"DEBIT MEMO":                 TransactionTypeDebitMemo,
	"WRITE OFF REVERSAL":         TransactionTypeWriteOffReversal,
	"FEE REDUCTION REVERSAL":     TransactionTypeFeeReductionReversal,
	"EXEMPTION":                  TransactionTypeExemption,
	"CREDIT EXEMPTION":           TransactionTypeExemption,
	"HARDSHIP":                   TransactionTypeHardship,
	"CREDIT HARDSHIP":            TransactionTypeHardship,
	"REMISSION":                  TransactionTypeRemission,
	"CREDIT REMISSION":           TransactionTypeRemission,
	"CREDIT REAPPLY":             TransactionTypeReapply,
	"MOTO CARD PAYMENT":          TransactionTypeMotoCardPayment,
	"ONLINE CARD PAYMENT":        TransactionTypeOnlineCardPayment,
	"SUPERVISION BACS PAYMENT":   TransactionTypeSupervisionBACSPayment,
	"OPG BACS PAYMENT":           TransactionTypeOPGBACSPayment,
	"DIRECT DEBIT PAYMENT":       TransactionTypeDirectDebitPayment,
	"SUPERVISION CHEQUE PAYMENT": TransactionTypeSupervisionChequePayment,
	"SOP UNALLOCATED":            TransactionTypeSOPUnallocatedPayment,
	"REFUND":                     TransactionTypeRefund,
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
	case TransactionTypeFeeReductionReversal:
		return "Fee reduction reversal"
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
	case TransactionTypeOPGBACSPayment:
		return "BACS payment (Main account)"
	case TransactionTypeDirectDebitPayment:
		return "Direct Debit payment"
	case TransactionTypeSupervisionChequePayment:
		return "Cheque payment"
	case TransactionTypeSOPUnallocatedPayment:
		return "SOP Unallocated"
	case TransactionTypeRefund:
		return "Refund"
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
	case TransactionTypeFeeReductionReversal:
		return "FEE REDUCTION REVERSAL"
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
	case TransactionTypeOPGBACSPayment:
		return "OPG BACS PAYMENT"
	case TransactionTypeDirectDebitPayment:
		return "DIRECT DEBIT PAYMENT"
	case TransactionTypeSupervisionChequePayment:
		return "SUPERVISION CHEQUE PAYMENT"
	case TransactionTypeSOPUnallocatedPayment:
		return "SOP UNALLOCATED"
	case TransactionTypeRefund:
		return "REFUND"
	default:
		return ""
	}
}

func (t TransactionType) IsPayment() bool {
	return slices.Contains(paymentTransactionTypes, t)
}

func (t TransactionType) IsReversalType() bool {
	return slices.Contains(reversalTransactionTypes, t)
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
