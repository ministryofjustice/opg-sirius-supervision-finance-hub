package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type InvoicesSchedule struct {
	Date         *shared.Date
	ScheduleType *shared.ScheduleType
}

const InvoicesScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	i.reference AS "Invoice reference",
	(i.amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
	TO_CHAR(i.raiseddate, 'YYYY-MM-DD') AS "Raised date"
	FROM supervision_finance.invoice i
	JOIN supervision_finance.finance_client fc ON fc.id = i.finance_client_id
	LEFT JOIN LATERAL (
    SELECT ifr.supervisionlevel AS supervision_level
    FROM supervision_finance.invoice_fee_range ifr
    WHERE ifr.invoice_id = i.id
    ORDER BY id DESC
    LIMIT 1
    ) sl ON $3 <> ''
	WHERE i.created_at::DATE = $1 AND i.feetype = $2 AND COALESCE(sl.supervision_level, '') = $3;
`

func (i *InvoicesSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Raised date",
	}
}

func (i *InvoicesSchedule) GetQuery() string {
	return InvoicesScheduleQuery
}

func (i *InvoicesSchedule) GetParams() []any {
	var (
		invoiceType      shared.InvoiceType
		supervisionLevel string
	)
	switch *i.ScheduleType {
	case shared.ScheduleTypeAdFeeInvoices:
		invoiceType = shared.InvoiceTypeAD
	case shared.ScheduleTypeS2FeeInvoices:
		invoiceType = shared.InvoiceTypeS2
	case shared.ScheduleTypeS3FeeInvoices:
		invoiceType = shared.InvoiceTypeS3
	case shared.ScheduleTypeB2FeeInvoices:
		invoiceType = shared.InvoiceTypeB2
	case shared.ScheduleTypeB3FeeInvoices:
		invoiceType = shared.InvoiceTypeB3
	case shared.ScheduleTypeSFFeeInvoicesGeneral:
		invoiceType = shared.InvoiceTypeSF
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeSFFeeInvoicesMinimal:
		invoiceType = shared.InvoiceTypeSF
		supervisionLevel = "MINIMAL"
	case shared.ScheduleTypeSEFeeInvoicesGeneral:
		invoiceType = shared.InvoiceTypeSE
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeSEFeeInvoicesMinimal:
		invoiceType = shared.InvoiceTypeSE
		supervisionLevel = "MINIMAL"
	case shared.ScheduleTypeSOFeeInvoicesGeneral:
		invoiceType = shared.InvoiceTypeSO
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeSOFeeInvoicesMinimal:
		invoiceType = shared.InvoiceTypeSO
		supervisionLevel = "MINIMAL"
	case shared.ScheduleTypeGAFeeInvoices:
		invoiceType = shared.InvoiceTypeGA
	case shared.ScheduleTypeGSFeeInvoices:
		invoiceType = shared.InvoiceTypeGS
	case shared.ScheduleTypeGTFeeInvoices:
		invoiceType = shared.InvoiceTypeGT
	default:
		invoiceType = shared.InvoiceTypeUnknown
	}

	return []any{i.Date.Time.Format("2006-01-02"), invoiceType.Key(), supervisionLevel}
}
