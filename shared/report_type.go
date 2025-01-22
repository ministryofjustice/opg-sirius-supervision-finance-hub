package shared

import "encoding/json"

var ReportAccountTypes = []ReportAccountType{
	ReportAccountTypeAgedDebt,
	ReportAccountTypeAgedDebtByCustomer,
	ReportAccountTypeUnappliedReceipts,
	ReportAccountTypeARPaidInvoiceReport,
	ReportAccountTypeTotalReceiptsReport,
	ReportAccountTypeBadDebtWriteOffReport,
	ReportAccountTypeFeeAccrual,
}

var reportAccountTypeMap = map[string]ReportAccountType{
	"AgedDebt":              ReportAccountTypeAgedDebt,
	"AgedDebtByCustomer":    ReportAccountTypeAgedDebtByCustomer,
	"UnappliedReceipts":     ReportAccountTypeUnappliedReceipts,
	"ARPaidInvoiceReport":   ReportAccountTypeARPaidInvoiceReport,
	"TotalReceiptsReport":   ReportAccountTypeTotalReceiptsReport,
	"BadDebtWriteOffReport": ReportAccountTypeBadDebtWriteOffReport,
	"FeeAccrual":            ReportAccountTypeFeeAccrual,
}

type ReportAccountType int

const (
	ReportAccountTypeUnknown ReportAccountType = iota
	ReportAccountTypeAgedDebt
	ReportAccountTypeAgedDebtByCustomer
	ReportAccountTypeUnappliedReceipts
	ReportAccountTypeARPaidInvoiceReport
	ReportAccountTypeTotalReceiptsReport
	ReportAccountTypeBadDebtWriteOffReport
	ReportAccountTypeFeeAccrual
)

func (i ReportAccountType) String() string {
	return i.Key()
}

func (i ReportAccountType) Translation() string {
	switch i {
	case ReportAccountTypeAgedDebt:
		return "Aged Debt"
	case ReportAccountTypeAgedDebtByCustomer:
		return "Ageing Buckets By Customer"
	case ReportAccountTypeUnappliedReceipts:
		return "Customer Credit Balance"
	case ReportAccountTypeARPaidInvoiceReport:
		return "AR Paid Invoice"
	case ReportAccountTypeTotalReceiptsReport:
		return "Total Receipts"
	case ReportAccountTypeBadDebtWriteOffReport:
		return "Bad Debt Write-off"
	case ReportAccountTypeFeeAccrual:
		return "Fee Accrual"
	default:
		return ""
	}
}

func (i ReportAccountType) Key() string {
	switch i {
	case ReportAccountTypeAgedDebt:
		return "AgedDebt"
	case ReportAccountTypeAgedDebtByCustomer:
		return "AgedDebtByCustomer"
	case ReportAccountTypeUnappliedReceipts:
		return "UnappliedReceipts"
	case ReportAccountTypeARPaidInvoiceReport:
		return "ARPaidInvoiceReport"
	case ReportAccountTypeTotalReceiptsReport:
		return "TotalReceiptsReport"
	case ReportAccountTypeBadDebtWriteOffReport:
		return "BadDebtWriteOffReport"
	case ReportAccountTypeFeeAccrual:
		return "FeeAccrual"
	default:
		return ""
	}
}

func ParseReportAccountType(s string) ReportAccountType {
	value, ok := reportAccountTypeMap[s]
	if !ok {
		return ReportAccountType(0)
	}
	return value
}

func (i ReportAccountType) Valid() bool {
	return i != ReportAccountTypeUnknown
}

func (i ReportAccountType) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Key())
}

func (i *ReportAccountType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*i = ParseReportAccountType(s)
	return nil
}
