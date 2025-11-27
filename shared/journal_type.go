package shared

import "encoding/json"

const (
	JournalTypeUnknown JournalType = iota
	JournalTypeReceiptTransactions
	JournalTypeNonReceiptTransactions
	JournalTypeUnappliedTransactions
	JournalTypeNonReceiptTransactionsHistoric
	JournalTypeReceiptTransactionsHistoric
)

var journalTypeMap = map[string]JournalType{
	"ReceiptTransactions":            JournalTypeReceiptTransactions,
	"NonReceiptTransactions":         JournalTypeNonReceiptTransactions,
	"RefundUnappliedTransactions":    JournalTypeUnappliedTransactions,
	"NonReceiptTransactionsHistoric": JournalTypeNonReceiptTransactionsHistoric,
	"ReceiptTransactionsHistoric":    JournalTypeReceiptTransactionsHistoric,
}

type JournalType int

func (j JournalType) String() string {
	return j.Key()
}

func (j JournalType) Translation() string {
	switch j {
	case JournalTypeReceiptTransactions:
		return "Receipt Transactions"
	case JournalTypeNonReceiptTransactions:
		return "Non Receipt Transactions"
	case JournalTypeUnappliedTransactions:
		return "Refunds & Unapplied Transactions"
	case JournalTypeNonReceiptTransactionsHistoric:
		return "Non Receipt Transactions (Historic)"
	case JournalTypeReceiptTransactionsHistoric:
		return "Receipt Transactions (Historic)"
	default:
		return ""
	}
}

func (j JournalType) Key() string {
	switch j {
	case JournalTypeReceiptTransactions:
		return "ReceiptTransactions"
	case JournalTypeNonReceiptTransactions:
		return "NonReceiptTransactions"
	case JournalTypeUnappliedTransactions:
		return "RefundUnappliedTransactions"
	case JournalTypeNonReceiptTransactionsHistoric:
		return "NonReceiptTransactionsHistoric"
	case JournalTypeReceiptTransactionsHistoric:
		return "ReceiptTransactionsHistoric"
	default:
		return ""
	}
}

func ParseJournalType(s string) JournalType {
	value, ok := journalTypeMap[s]
	if !ok {
		return JournalType(0)
	}
	return value
}

func (j JournalType) Valid() bool {
	return j != JournalTypeUnknown
}

func (j JournalType) MarshalJSON() ([]byte, error) {
	return json.Marshal(j.Key())
}

func (j *JournalType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*j = ParseJournalType(s)
	return nil
}
