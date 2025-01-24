package shared

import "encoding/json"

var reportAccountTypeMap = map[string]ReportType{
	"AgedDebt":           ReportTypeAgedDebt,
	"AgedDebtByCustomer": ReportTypeAgedDebtByCustomer,
	"UnappliedReceipts":  ReportTypeUnappliedReceipts,
	"ARPaidInvoice":      ReportTypeARPaidInvoice,
	"TotalReceipts":      ReportTypeTotalReceipts,
	"BadDebtWriteOff":    ReportTypeBadDebtWriteOff,
	"FeeAccrual":         ReportTypeFeeAccrual,
}

type ReportType int

const (
	ReportAccountTypeUnknown ReportType = iota
	ReportTypeAgedDebt
	ReportTypeAgedDebtByCustomer
	ReportTypeUnappliedReceipts
	ReportTypeARPaidInvoice
	ReportTypeTotalReceipts
	ReportTypeBadDebtWriteOff
	ReportTypeFeeAccrual
)

func (i ReportType) String() string {
	return i.Key()
}

func (i ReportType) Translation() string {
	switch i {
	case ReportTypeAgedDebt:
		return "Aged Debt"
	case ReportTypeAgedDebtByCustomer:
		return "Ageing Buckets By Customer"
	case ReportTypeUnappliedReceipts:
		return "Customer Credit Balance"
	case ReportTypeARPaidInvoice:
		return "AR Paid Invoice"
	case ReportTypeTotalReceipts:
		return "Total Receipts"
	case ReportTypeBadDebtWriteOff:
		return "Bad Debt Write-off"
	case ReportTypeFeeAccrual:
		return "Fee Accrual"
	default:
		return ""
	}
}

func (i ReportType) Key() string {
	switch i {
	case ReportTypeAgedDebt:
		return "AgedDebt"
	case ReportTypeAgedDebtByCustomer:
		return "AgedDebtByCustomer"
	case ReportTypeUnappliedReceipts:
		return "UnappliedReceipts"
	case ReportTypeARPaidInvoice:
		return "ARPaidInvoice"
	case ReportTypeTotalReceipts:
		return "TotalReceipts"
	case ReportTypeBadDebtWriteOff:
		return "BadDebtWriteOff"
	case ReportTypeFeeAccrual:
		return "FeeAccrual"
	default:
		return ""
	}
}

func ParseReportAccountType(s string) ReportType {
	value, ok := reportAccountTypeMap[s]
	if !ok {
		return ReportType(0)
	}
	return value
}

func (i ReportType) Valid() bool {
	return i != ReportAccountTypeUnknown
}

func (i ReportType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportAccountType(s)
	return nil
}
