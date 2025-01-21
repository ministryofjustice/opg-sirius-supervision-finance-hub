package shared

import "encoding/json"

var ReportAccountTypes = []ReportAccountType{
	ReportAccountTypeAgedDebt,
	ReportAccountTypeAgedDebtByCustomer,
	ReportAccountTypeUnappliedReceipts,
	ReportAccountTypeCustomerAgeingBuckets,
	ReportAccountTypeARPaidInvoiceReport,
	ReportAccountTypePaidInvoiceTransactionLines,
	ReportAccountTypeTotalReceiptsReport,
	ReportAccountTypeBadDebtWriteOffReport,
	ReportAccountTypeFeeAccrual,
	ReportAccountTypeInvoiceAdjustments,
}

var reportAccountTypeMap = map[string]ReportAccountType{
	"AgedDebt":                    ReportAccountTypeAgedDebt,
	"AgedDebtByCustomer":          ReportAccountTypeAgedDebtByCustomer,
	"UnappliedReceipts":           ReportAccountTypeUnappliedReceipts,
	"CustomerAgeingBuckets":       ReportAccountTypeCustomerAgeingBuckets,
	"ARPaidInvoiceReport":         ReportAccountTypeARPaidInvoiceReport,
	"PaidInvoiceTransactionLines": ReportAccountTypePaidInvoiceTransactionLines,
	"TotalReceiptsReport":         ReportAccountTypeTotalReceiptsReport,
	"BadDebtWriteOffReport":       ReportAccountTypeBadDebtWriteOffReport,
	"FeeAccrual":                  ReportAccountTypeFeeAccrual,
	"InvoiceAdjustments":          ReportAccountTypeInvoiceAdjustments,
}

type ReportAccountType int

const (
	ReportAccountTypeUnknown ReportAccountType = iota
	ReportAccountTypeAgedDebt
	ReportAccountTypeAgedDebtByCustomer
	ReportAccountTypeUnappliedReceipts
	ReportAccountTypeCustomerAgeingBuckets
	ReportAccountTypeARPaidInvoiceReport
	ReportAccountTypePaidInvoiceTransactionLines
	ReportAccountTypeTotalReceiptsReport
	ReportAccountTypeBadDebtWriteOffReport
	ReportAccountTypeFeeAccrual
	ReportAccountTypeInvoiceAdjustments
)

func (i ReportAccountType) String() string {
	return i.Key()
}

func (i ReportAccountType) Translation() string {
	switch i {
	case ReportAccountTypeAgedDebt:
		return "Aged Debt"
	case ReportAccountTypeAgedDebtByCustomer:
		return "Aged Debt By Customer"
	case ReportAccountTypeUnappliedReceipts:
		return "Unapplied Receipts"
	case ReportAccountTypeCustomerAgeingBuckets:
		return "Customer Ageing Buckets"
	case ReportAccountTypeARPaidInvoiceReport:
		return "AR Paid Invoice Report"
	case ReportAccountTypePaidInvoiceTransactionLines:
		return "Paid Invoice Transaction Lines"
	case ReportAccountTypeTotalReceiptsReport:
		return "Total Receipts Report"
	case ReportAccountTypeBadDebtWriteOffReport:
		return "Bad Debt Write-off Report"
	case ReportAccountTypeFeeAccrual:
		return "Fee Accrual"
	case ReportAccountTypeInvoiceAdjustments:
		return "Invoice Adjustments"
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
	case ReportAccountTypeCustomerAgeingBuckets:
		return "CustomerAgeingBuckets"
	case ReportAccountTypeARPaidInvoiceReport:
		return "ARPaidInvoiceReport"
	case ReportAccountTypePaidInvoiceTransactionLines:
		return "PaidInvoiceTransactionLines"
	case ReportAccountTypeTotalReceiptsReport:
		return "TotalReceiptsReport"
	case ReportAccountTypeBadDebtWriteOffReport:
		return "BadDebtWriteOffReport"
	case ReportAccountTypeFeeAccrual:
		return "FeeAccrual"
	case ReportAccountTypeInvoiceAdjustments:
		return "InvoiceAdjustments"
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
