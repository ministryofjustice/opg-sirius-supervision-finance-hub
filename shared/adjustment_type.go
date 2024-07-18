package shared

import (
	"encoding/json"
)

type Valid interface {
	Valid() bool
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

func (i AdjustmentType) Translation() string {
	switch i {
	case AdjustmentTypeWriteOff:
		return "Write off"
	case AdjustmentTypeCreditMemo:
		return "Add credit"
	case AdjustmentTypeDebitMemo:
		return "Add debit"
	case AdjustmentTypeWriteOffReversal:
		return "Write off reversal (Not yet implemented)"
	default:
		return ""
	}
}

func (i AdjustmentType) Key() string {
	switch i {
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

func (i AdjustmentType) AmountRequired() bool {
	switch i {
	case AdjustmentTypeCreditMemo, AdjustmentTypeDebitMemo:
		return true
	default:
		return false
	}
}

func (i AdjustmentType) Valid() bool {
	return i != AdjustmentTypeUnknown
}

func ParseAdjustmentType(s string) AdjustmentType {
	value, ok := adjustmentTypeMap[s]
	if !ok {
		return AdjustmentType(0)
	}
	return value
}

func (i AdjustmentType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *AdjustmentType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseAdjustmentType(s)
	return nil
}
