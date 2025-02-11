package shared

import "encoding/json"

var ReportJournalTypes = []ReportJournalType{
	ReportTypeReceiptTransactions,
	ReportTypeNonReceiptTransactions,
}

const (
	ReportTypeUnknown ReportJournalType = iota
	ReportTypeReceiptTransactions
	ReportTypeNonReceiptTransactions
)

var reportJournalTypeMap = map[string]ReportJournalType{
	"ReceiptTransactions":    ReportTypeReceiptTransactions,
	"NonReceiptTransactions": ReportTypeNonReceiptTransactions,
}

type ReportJournalType int

func (i ReportJournalType) String() string {
	return i.Key()
}

func (i ReportJournalType) Translation() string {
	switch i {
	case ReportTypeReceiptTransactions:
		return "Receipt Transactions"
	case ReportTypeNonReceiptTransactions:
		return "Non Receipt Transactions"
	default:
		return ""
	}
}

func (i ReportJournalType) Key() string {
	switch i {
	case ReportTypeReceiptTransactions:
		return "ReceiptTransactions"
	case ReportTypeNonReceiptTransactions:
		return "NonReceiptTransactions"
	default:
		return ""
	}
}

func ParseReportJournalType(s string) ReportJournalType {
	value, ok := reportJournalTypeMap[s]
	if !ok {
		return ReportJournalType(0)
	}
	return value
}

func (i ReportJournalType) Valid() bool {
	return i != ReportTypeUnknown
}

func (i ReportJournalType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportJournalType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportJournalType(s)
	return nil
}
