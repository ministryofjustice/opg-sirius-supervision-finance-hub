package shared

import (
	"encoding/json"
)

type Enum interface {
	Valid() bool
	Key() string
}

type AdjustmentType int

const (
	AdjustmentTypeUnknown AdjustmentType = iota
	AdjustmentTypeWriteOff
	AdjustmentTypeCreditMemo
	AdjustmentTypeDebitMemo
	AdjustmentTypeWriteOffReversal
)

var adjustmentTypeMap = map[string]AdjustmentType{
	"CREDIT WRITE OFF":   AdjustmentTypeWriteOff,
	"CREDIT MEMO":        AdjustmentTypeCreditMemo,
	"DEBIT MEMO":         AdjustmentTypeDebitMemo,
	"WRITE OFF REVERSAL": AdjustmentTypeWriteOffReversal,
}

func (a AdjustmentType) String() string {
	switch a {
	case AdjustmentTypeWriteOff:
		return "Write off"
	case AdjustmentTypeCreditMemo:
		return "Credit memo"
	case AdjustmentTypeDebitMemo:
		return "Debit memo"
	case AdjustmentTypeWriteOffReversal:
		return "Write off reversal"
	default:
		return ""
	}
}

func (a AdjustmentType) Translation() string {
	switch a {
	case AdjustmentTypeWriteOff:
		return "Write off"
	case AdjustmentTypeCreditMemo:
		return "Add credit"
	case AdjustmentTypeDebitMemo:
		return "Add debit"
	case AdjustmentTypeWriteOffReversal:
		return "Write off reversal"
	default:
		return ""
	}
}

func (a AdjustmentType) Key() string {
	switch a {
	case AdjustmentTypeWriteOff:
		return "CREDIT WRITE OFF"
	case AdjustmentTypeCreditMemo:
		return "CREDIT MEMO"
	case AdjustmentTypeDebitMemo:
		return "DEBIT MEMO"
	case AdjustmentTypeWriteOffReversal:
		return "WRITE OFF REVERSAL"
	default:
		return ""
	}
}

func (a AdjustmentType) AmountRequired() bool {
	switch a {
	case AdjustmentTypeCreditMemo, AdjustmentTypeDebitMemo:
		return true
	default:
		return false
	}
}

func (a AdjustmentType) CanOverride() bool {
	return a == AdjustmentTypeWriteOffReversal
}

func (a AdjustmentType) Valid() bool {
	return a != AdjustmentTypeUnknown
}

func ParseAdjustmentType(s string) AdjustmentType {
	value, ok := adjustmentTypeMap[s]
	if !ok {
		return AdjustmentType(0)
	}
	return value
}

func (a AdjustmentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Key())
}

func (a *AdjustmentType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = ParseAdjustmentType(s)
	return nil
}
