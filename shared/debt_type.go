package shared

import "encoding/json"

var ReportDebtTypes = []ReportDebtType{
	ReportDebtTypeFeeChase,
	ReportDebtTypeFinalFee,
}

type ReportDebtType int

const (
	ReportDebtTypeUnknown ReportDebtType = iota
	ReportDebtTypeFeeChase
	ReportDebtTypeFinalFee
)

var reportDebtTypeMap = map[string]ReportDebtType{
	"FeeChase": ReportDebtTypeFeeChase,
	"FinalFee": ReportDebtTypeFinalFee,
}

func (i ReportDebtType) String() string {
	return i.Key()
}

func (i ReportDebtType) Translation() string {
	switch i {
	case ReportDebtTypeFeeChase:
		return "Fee Chase"
	case ReportDebtTypeFinalFee:
		return "Final Fee"
	default:
		return ""
	}
}

func (i ReportDebtType) Key() string {
	switch i {
	case ReportDebtTypeFeeChase:
		return "FeeChase"
	case ReportDebtTypeFinalFee:
		return "FinalFee"
	default:
		return ""
	}
}

func ParseReportDebtType(s string) ReportDebtType {
	value, ok := reportDebtTypeMap[s]
	if !ok {
		return ReportDebtType(0)
	}
	return value
}

func (i ReportDebtType) Valid() bool {
	return i != ReportDebtTypeUnknown
}

func (i ReportDebtType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportDebtType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportDebtType(s)
	return nil
}
