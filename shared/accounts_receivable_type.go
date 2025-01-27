package shared

import "encoding/json"

var ReportAccountsReceivableTypes = []ReportAccountsReceivableType{
	ReportAccountsReceivableTypeAgedDebt,
	ReportAccountsReceivableTypeAgedDebtByCustomer,
	ReportAccountsReceivableTypeUnappliedReceipts,
	ReportAccountsReceivableTypeARPaidInvoice,
	ReportAccountsReceivableTypeTotalReceipts,
	ReportAccountsReceivableTypeBadDebtWriteOff,
	ReportAccountsReceivableTypeFeeAccrual,
}

var ReportAccountsReceivableTypeMap = map[string]ReportAccountsReceivableType{
	"AgedDebt":           ReportAccountsReceivableTypeAgedDebt,
	"AgedDebtByCustomer": ReportAccountsReceivableTypeAgedDebtByCustomer,
	"UnappliedReceipts":  ReportAccountsReceivableTypeUnappliedReceipts,
	"ARPaidInvoice":      ReportAccountsReceivableTypeARPaidInvoice,
	"TotalReceipts":      ReportAccountsReceivableTypeTotalReceipts,
	"BadDebtWriteOff":    ReportAccountsReceivableTypeBadDebtWriteOff,
	"FeeAccrual":         ReportAccountsReceivableTypeFeeAccrual,
}

type ReportAccountsReceivableType int

const (
	ReportAccountsReceivableTypeUnknown ReportAccountsReceivableType = iota
	ReportAccountsReceivableTypeAgedDebt
	ReportAccountsReceivableTypeAgedDebtByCustomer
	ReportAccountsReceivableTypeUnappliedReceipts
	ReportAccountsReceivableTypeARPaidInvoice
	ReportAccountsReceivableTypeTotalReceipts
	ReportAccountsReceivableTypeBadDebtWriteOff
	ReportAccountsReceivableTypeFeeAccrual
)

func (i ReportAccountsReceivableType) String() string {
	return i.Key()
}

func (i ReportAccountsReceivableType) Translation() string {
	switch i {
	case ReportAccountsReceivableTypeAgedDebt:
		return "Aged Debt"
	case ReportAccountsReceivableTypeAgedDebtByCustomer:
		return "Ageing Buckets By Customer"
	case ReportAccountsReceivableTypeUnappliedReceipts:
		return "Customer Credit Balance"
	case ReportAccountsReceivableTypeARPaidInvoice:
		return "AR Paid Invoice"
	case ReportAccountsReceivableTypeTotalReceipts:
		return "Total Receipts"
	case ReportAccountsReceivableTypeBadDebtWriteOff:
		return "Bad Debt Write-off"
	case ReportAccountsReceivableTypeFeeAccrual:
		return "Fee Accrual"
	default:
		return ""
	}
}

func (i ReportAccountsReceivableType) Key() string {
	switch i {
	case ReportAccountsReceivableTypeAgedDebt:
		return "AgedDebt"
	case ReportAccountsReceivableTypeAgedDebtByCustomer:
		return "AgedDebtByCustomer"
	case ReportAccountsReceivableTypeUnappliedReceipts:
		return "UnappliedReceipts"
	case ReportAccountsReceivableTypeARPaidInvoice:
		return "ARPaidInvoice"
	case ReportAccountsReceivableTypeTotalReceipts:
		return "TotalReceipts"
	case ReportAccountsReceivableTypeBadDebtWriteOff:
		return "BadDebtWriteOff"
	case ReportAccountsReceivableTypeFeeAccrual:
		return "FeeAccrual"
	default:
		return ""
	}
}

func ParseReportAccountsReceivableType(s string) ReportAccountsReceivableType {
	value, ok := ReportAccountsReceivableTypeMap[s]
	if !ok {
		return ReportAccountsReceivableType(0)
	}
	return value
}

func (i ReportAccountsReceivableType) Valid() bool {
	return i != ReportAccountsReceivableTypeUnknown
}

func (i ReportAccountsReceivableType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportAccountsReceivableType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportAccountsReceivableType(s)
	return nil
}
