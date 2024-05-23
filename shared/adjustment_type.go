package shared

import (
	"encoding/json"
)

type Valid interface {
	Valid() bool
}

var AdjustmentTypes = []AdjustmentType{
	AdjustmentTypeWriteOff,
	AdjustmentTypeAddCredit,
	AdjustmentTypeAddDebit,
}

type AdjustmentType int

const (
	AdjustmentTypeUnknown AdjustmentType = iota
	AdjustmentTypeWriteOff
	AdjustmentTypeAddCredit
	AdjustmentTypeAddDebit
)

var adjustmentTypeMap = map[string]AdjustmentType{
	"CREDIT WRITE OFF": AdjustmentTypeWriteOff,
	"CREDIT_WRITE_OFF": AdjustmentTypeWriteOff,
	"CREDIT MEMO":      AdjustmentTypeAddCredit,
	"CREDIT_MEMO":      AdjustmentTypeAddCredit,
	"UNKNOWN DEBIT":    AdjustmentTypeAddDebit,
	"UNKNOWN_DEBIT":    AdjustmentTypeAddDebit,
}

func (i AdjustmentType) Translation() string {
	switch i {
	case AdjustmentTypeWriteOff:
		return "Write off"
	case AdjustmentTypeAddCredit:
		return "Add credit"
	case AdjustmentTypeAddDebit:
		return "Add debit"
	default:
		return ""
	}
}

func (i AdjustmentType) Key() string {
	switch i {
	case AdjustmentTypeWriteOff:
		return "CREDIT_WRITE_OFF"
	case AdjustmentTypeAddCredit:
		return "CREDIT_MEMO"
	case AdjustmentTypeAddDebit:
		return "UNKNOWN_DEBIT"
	default:
		return ""
	}
}

func (i AdjustmentType) DbValue() string {
	switch i {
	case AdjustmentTypeWriteOff:
		return "CREDIT WRITE OFF"
	case AdjustmentTypeAddCredit:
		return "CREDIT MEMO"
	case AdjustmentTypeAddDebit:
		return "UNKNOWN DEBIT"
	default:
		return ""
	}
}

func (i AdjustmentType) AmountRequired() bool {
	switch i {
	case AdjustmentTypeAddCredit, AdjustmentTypeAddDebit:
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
