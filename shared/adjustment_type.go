package shared

import (
	"encoding/json"
)

var AdjustmentTypes = []AdjustmentType{
	AdjustmentTypeWriteOff,
	AdjustmentTypeAddCredit,
	AdjustmentTypeAddDebit,
	AdjustmentTypeUnapply,
	AdjustmentTypeReapply,
}

type AdjustmentType int

const (
	AdjustmentTypeUnknown AdjustmentType = iota
	AdjustmentTypeWriteOff
	AdjustmentTypeAddCredit
	AdjustmentTypeAddDebit
	AdjustmentTypeUnapply
	AdjustmentTypeReapply
)

var adjustmentTypeMap = map[string]AdjustmentType{
	"CREDIT WRITE OFF": AdjustmentTypeWriteOff,
	"CREDIT MEMO":      AdjustmentTypeAddCredit,
	"UNKNOWN DEBIT":    AdjustmentTypeAddDebit,
	"UNAPPLY":          AdjustmentTypeUnapply,
	"REAPPLY":          AdjustmentTypeReapply,
}

func (i AdjustmentType) Translation() string {
	switch i {
	case AdjustmentTypeWriteOff:
		return "Write off"
	case AdjustmentTypeAddCredit:
		return "Add credit"
	case AdjustmentTypeAddDebit:
		return "Add debit"
	case AdjustmentTypeUnapply:
		return "Unapply"
	case AdjustmentTypeReapply:
		return "Reapply"
	default:
		return ""
	}
}

func (i AdjustmentType) Key() string {
	switch i {
	case AdjustmentTypeWriteOff:
		return "CREDIT WRITE OFF"
	case AdjustmentTypeAddCredit:
		return "CREDIT MEMO"
	case AdjustmentTypeAddDebit:
		return "UNKNOWN DEBIT"
	case AdjustmentTypeUnapply:
		return "UNAPPLY"
	case AdjustmentTypeReapply:
		return "REAPPLY"
	default:
		return ""
	}
}

func (i AdjustmentType) AmountRequired() bool {
	switch i {
	case AdjustmentTypeAddCredit, AdjustmentTypeAddDebit, AdjustmentTypeUnapply, AdjustmentTypeReapply:
		return true
	default:
		return false
	}
}

func (i AdjustmentType) IsValid() bool {
	switch i {
	case AdjustmentTypeAddCredit, AdjustmentTypeAddDebit, AdjustmentTypeUnapply, AdjustmentTypeReapply:
		return true
	default:
		return false
	}
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
