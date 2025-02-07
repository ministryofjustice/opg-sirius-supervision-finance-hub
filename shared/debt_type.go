package shared

import "encoding/json"

type DebtType int

const (
	DebtTypeUnknown DebtType = iota
	DebtTypeFeeChase
	DebtTypeFinalFee
)

var debtTypeMap = map[string]DebtType{
	"FeeChase": DebtTypeFeeChase,
	"FinalFee": DebtTypeFinalFee,
}

func (d DebtType) String() string {
	return d.Key()
}

func (d DebtType) Translation() string {
	switch d {
	case DebtTypeFeeChase:
		return "Fee Chase"
	case DebtTypeFinalFee:
		return "Final Fee"
	default:
		return ""
	}
}

func (d DebtType) Key() string {
	switch d {
	case DebtTypeFeeChase:
		return "FeeChase"
	case DebtTypeFinalFee:
		return "FinalFee"
	default:
		return ""
	}
}

func ParseReportDebtType(s string) DebtType {
	value, ok := debtTypeMap[s]
	if !ok {
		return DebtType(0)
	}
	return value
}

func (d DebtType) Valid() bool {
	return d != DebtTypeUnknown
}

func (d DebtType) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Key())
}

func (d *DebtType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*d = ParseReportDebtType(s)
	return nil
}
