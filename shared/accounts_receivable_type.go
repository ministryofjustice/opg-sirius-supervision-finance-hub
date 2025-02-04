package shared

import "encoding/json"

var AccountsReceivableTypeMap = map[string]AccountsReceivableType{
	"AgedDebt":           AccountsReceivableTypeAgedDebt,
	"AgedDebtByCustomer": AccountsReceivableTypeAgedDebtByCustomer,
	"UnappliedReceipts":  AccountsReceivableTypeUnappliedReceipts,
	"ARPaidInvoice":      AccountsReceivableTypeARPaidInvoice,
	"TotalReceipts":      AccountsReceivableTypeTotalReceipts,
	"BadDebtWriteOff":    AccountsReceivableTypeBadDebtWriteOff,
	"FeeAccrual":         AccountsReceivableTypeFeeAccrual,
	"InvoiceAdjustments": AccountsReceivableTypeInvoiceAdjustments,
}

type AccountsReceivableType int

const (
	AccountsReceivableTypeUnknown AccountsReceivableType = iota
	AccountsReceivableTypeAgedDebt
	AccountsReceivableTypeAgedDebtByCustomer
	AccountsReceivableTypeUnappliedReceipts
	AccountsReceivableTypeARPaidInvoice
	AccountsReceivableTypeTotalReceipts
	AccountsReceivableTypeBadDebtWriteOff
	AccountsReceivableTypeFeeAccrual
	AccountsReceivableTypeInvoiceAdjustments
)

func (a AccountsReceivableType) String() string {
	return a.Key()
}

func (a AccountsReceivableType) Translation() string {
	switch a {
	case AccountsReceivableTypeAgedDebt:
		return "Aged Debt"
	case AccountsReceivableTypeAgedDebtByCustomer:
		return "Ageing Buckets By Customer"
	case AccountsReceivableTypeUnappliedReceipts:
		return "Customer Credit Balance"
	case AccountsReceivableTypeARPaidInvoice:
		return "AR Paid Invoice"
	case AccountsReceivableTypeTotalReceipts:
		return "Total Receipts"
	case AccountsReceivableTypeBadDebtWriteOff:
		return "Bad Debt Write-off"
	case AccountsReceivableTypeFeeAccrual:
		return "Fee Accrual"
	case AccountsReceivableTypeInvoiceAdjustments:
		return "Invoice Adjustments"
	default:
		return ""
	}
}

func (a AccountsReceivableType) Key() string {
	switch a {
	case AccountsReceivableTypeAgedDebt:
		return "AgedDebt"
	case AccountsReceivableTypeAgedDebtByCustomer:
		return "AgedDebtByCustomer"
	case AccountsReceivableTypeUnappliedReceipts:
		return "UnappliedReceipts"
	case AccountsReceivableTypeARPaidInvoice:
		return "ARPaidInvoice"
	case AccountsReceivableTypeTotalReceipts:
		return "TotalReceipts"
	case AccountsReceivableTypeBadDebtWriteOff:
		return "BadDebtWriteOff"
	case AccountsReceivableTypeFeeAccrual:
		return "FeeAccrual"
	case AccountsReceivableTypeInvoiceAdjustments:
		return "InvoiceAdjustments"
	default:
		return ""
	}
}

func ParseAccountsReceivableType(s string) AccountsReceivableType {
	value, ok := AccountsReceivableTypeMap[s]
	if !ok {
		return AccountsReceivableType(0)
	}
	return value
}

func (a AccountsReceivableType) MarshalJSON() ([]byte, error) {
	return json.Marshal(a.Key())
}

func (a *AccountsReceivableType) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	*a = ParseAccountsReceivableType(s)
	return nil
}

func (a AccountsReceivableType) Valid() bool {
	return a != AccountsReceivableTypeUnknown
}
