package shared

import "encoding/json"

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

func (r ReportsType) String() string {
	return r.Key()
}

func (r ReportsType) Translation() string {
	switch r {
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

func (r ReportsType) Key() string {
	switch r {
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

func (r ReportsType) Valid() bool {
	return r != ReportsTypeUnknown
}

func (r ReportsType) RequiresDateValidation() bool {
	switch r {
	case ReportsTypeJournal:
		return true
	}
	return false
}

func (r ReportsType) MarshalJSON() ([]byte, error) {
	return json.Marshal(r.Key())
}

func (r *ReportsType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*r = ParseReportsType(s)
	return nil
}
