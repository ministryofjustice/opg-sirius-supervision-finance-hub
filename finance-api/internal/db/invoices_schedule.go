package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

// InvoicesSchedule generates a report of invoices raised for a given date and invoice type.
// This is used by the Billing team to reconcile invoice debt raised, and each invoice schedule should correlate to a line
// in the non-receipts transactions journal by line description.
type InvoicesSchedule struct {
	ReportQuery
	InvoicesScheduleInput
}

type InvoicesScheduleInput struct {
	Date         *shared.Date
	ScheduleType *shared.ScheduleType
}

func NewInvoicesSchedule(input InvoicesScheduleInput) ReportQuery {
	return &InvoicesSchedule{
		ReportQuery:           NewReportQuery(InvoicesScheduleQuery),
		InvoicesScheduleInput: input,
	}
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
    ) sl ON TRUE
WHERE i.created_at::DATE = $1
  AND i.feetype = $2
  AND (
      ($3 = '' AND sl.supervision_level IS NULL)
      OR ($3 <> '' AND sl.supervision_level = $3)
  );
`

func (i *InvoicesSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Raised date",
	}
}

func (i *InvoicesSchedule) GetParams() []any {
	var (
		invoiceType      shared.InvoiceType
		supervisionLevel string
	)
	switch *i.ScheduleType {
	case shared.ScheduleTypeS2FeeInvoices,
		shared.ScheduleTypeB2FeeInvoices,
		shared.ScheduleTypeSFFeeInvoicesGeneral,
		shared.ScheduleTypeSEFeeInvoicesGeneral,
		shared.ScheduleTypeSOFeeInvoicesGeneral:
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeS3FeeInvoices,
		shared.ScheduleTypeB3FeeInvoices,
		shared.ScheduleTypeSFFeeInvoicesMinimal,
		shared.ScheduleTypeSEFeeInvoicesMinimal,
		shared.ScheduleTypeSOFeeInvoicesMinimal:
		supervisionLevel = "MINIMAL"
	default:
		supervisionLevel = ""
	}

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
	case shared.ScheduleTypeSFFeeInvoicesMinimal:
		invoiceType = shared.InvoiceTypeSF
	case shared.ScheduleTypeSEFeeInvoicesGeneral:
		invoiceType = shared.InvoiceTypeSE
	case shared.ScheduleTypeSEFeeInvoicesMinimal:
		invoiceType = shared.InvoiceTypeSE
	case shared.ScheduleTypeSOFeeInvoicesGeneral:
		invoiceType = shared.InvoiceTypeSO
	case shared.ScheduleTypeSOFeeInvoicesMinimal:
		invoiceType = shared.InvoiceTypeSO
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
