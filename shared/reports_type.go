package shared

import "encoding/json"

var ReportsTypes = []ReportsType{
	ReportsTypeJournal,
	ReportsTypeSchedule,
	ReportsTypeAccountsReceivable,
	ReportsTypeDebt,
}

type ReportsType int

const (
	ReportsTypeUnknown ReportsType = iota
	ReportsTypeJournal
	ReportsTypeSchedule
	ReportsTypeAccountsReceivable
	ReportsTypeDebt
)

var reportsTypeMap = map[string]ReportsType{
	"Journal":            ReportsTypeJournal,
	"Schedule":           ReportsTypeSchedule,
	"AccountsReceivable": ReportsTypeAccountsReceivable,
	"Debt":               ReportsTypeDebt,
}

func (i ReportsType) String() string {
	return i.Key()
}

func (i ReportsType) Translation() string {
	switch i {
	case ReportsTypeJournal:
		return "Journal"
	case ReportsTypeSchedule:
		return "Schedule"
	case ReportsTypeAccountsReceivable:
		return "Accounts Receivable"
	case ReportsTypeDebt:
		return "Debt"
	default:
		return ""
	}
}

func (i ReportsType) Key() string {
	switch i {
	case ReportsTypeJournal:
		return "Journal"
	case ReportsTypeSchedule:
		return "Schedule"
	case ReportsTypeAccountsReceivable:
		return "AccountsReceivable"
	case ReportsTypeDebt:
		return "Debt"
	default:
		return ""
	}
}

func ParseReportsType(s string) ReportsType {
	value, ok := reportsTypeMap[s]
	if !ok {
		return ReportsType(0)
	}
	return value
}

func (i ReportsType) Valid() bool {
	return i != ReportsTypeUnknown
}

func (i ReportsType) RequiresDateValidation() bool {
	switch i {
	case ReportsTypeJournal:
		return true
	}
	return false
}

func (i ReportsType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportsType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportsType(s)
	return nil
}
