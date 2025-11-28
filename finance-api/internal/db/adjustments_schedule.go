package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

// AdjustmentsSchedule generates a report of all invoice adjustments for a given date and adjustment type.
// This is used by the Billing team to reconcile adjustments, and each adjustment schedule should correlate to a line
// in the non-receipts transactions journal by line description.
type AdjustmentsSchedule struct {
	ReportQuery
	AdjustmentsScheduleInput
}

type AdjustmentsScheduleInput struct {
	Date         *shared.Date
	ScheduleType *shared.ScheduleType
}

func NewAdjustmentsSchedule(input AdjustmentsScheduleInput) ReportQuery {
	return &AdjustmentsSchedule{
		ReportQuery:              NewReportQuery(AdjustmentsScheduleQuery),
		AdjustmentsScheduleInput: input,
	}
}

const AdjustmentsScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	i.reference AS "Invoice reference",
	(CASE WHEN $5 THEN la.amount ELSE ABS(la.amount) END / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
	TO_CHAR(l.created_at, 'YYYY-MM-DD') AS "Created date"
	FROM supervision_finance.ledger l
	    JOIN supervision_finance.ledger_allocation la ON l.id = la.ledger_id AND la.status = 'ALLOCATED'
	    JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
	    JOIN supervision_finance.invoice i ON i.id = la.invoice_id
		LEFT JOIN LATERAL (
			SELECT ifr.supervisionlevel AS supervision_level
			FROM supervision_finance.invoice_fee_range ifr
			WHERE ifr.invoice_id = i.id
			ORDER BY id DESC
			LIMIT 1
		) sl ON TRUE
	WHERE l.created_at::DATE = $1
	  AND l.type = ANY($2)
	  AND (
		  ($3 = '' AND sl.supervision_level IS NULL)
		  OR ($3 <> '' AND sl.supervision_level = $3)
  	) AND i.feetype = ANY($4);
`

func (a *AdjustmentsSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Created date",
	}
}

func (a *AdjustmentsSchedule) GetParams() []any {
	var (
		ledgerTypes      []string
		supervisionLevel string
		invoiceTypes     []string
		includeNegatives = false
	)
	switch *a.ScheduleType {
	case shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeGeneralWriteOffs,
		shared.ScheduleTypeGeneralWriteOffReversals,
		shared.ScheduleTypeGeneralFeeReductionReversals:
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeMinimalFeeReductions,
		shared.ScheduleTypeMinimalManualCredits,
		shared.ScheduleTypeMinimalManualDebits,
		shared.ScheduleTypeMinimalWriteOffs,
		shared.ScheduleTypeMinimalWriteOffReversals,
		shared.ScheduleTypeMinimalFeeReductionReversals:
		supervisionLevel = "MINIMAL"
	default:
		supervisionLevel = ""
	}

	switch *a.ScheduleType {
	case shared.ScheduleTypeADFeeReductions,
		shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeMinimalFeeReductions,
		shared.ScheduleTypeGAFeeReductions,
		shared.ScheduleTypeGSFeeReductions,
		shared.ScheduleTypeGTFeeReductions:
		ledgerTypes = []string{
			"CREDIT " + shared.TransactionTypeHardship.Key(),
			"CREDIT " + shared.TransactionTypeExemption.Key(),
			"CREDIT " + shared.TransactionTypeRemission.Key(),
		}
	case shared.ScheduleTypeADManualCredits,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeMinimalManualCredits,
		shared.ScheduleTypeGAManualCredits,
		shared.ScheduleTypeGSManualCredits,
		shared.ScheduleTypeGTManualCredits:
		ledgerTypes = []string{
			shared.TransactionTypeCreditMemo.Key(),
		}
	case shared.ScheduleTypeADManualDebits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeMinimalManualDebits,
		shared.ScheduleTypeGAManualDebits,
		shared.ScheduleTypeGSManualDebits,
		shared.ScheduleTypeGTManualDebits:
		ledgerTypes = []string{
			shared.TransactionTypeDebitMemo.Key(),
		}
	case shared.ScheduleTypeADWriteOffs,
		shared.ScheduleTypeGeneralWriteOffs,
		shared.ScheduleTypeMinimalWriteOffs,
		shared.ScheduleTypeGAWriteOffs,
		shared.ScheduleTypeGSWriteOffs,
		shared.ScheduleTypeGTWriteOffs:
		ledgerTypes = []string{
			shared.TransactionTypeWriteOff.Key(),
		}
	case shared.ScheduleTypeADWriteOffReversals,
		shared.ScheduleTypeGeneralWriteOffReversals,
		shared.ScheduleTypeMinimalWriteOffReversals,
		shared.ScheduleTypeGAWriteOffReversals,
		shared.ScheduleTypeGSWriteOffReversals,
		shared.ScheduleTypeGTWriteOffReversals:
		ledgerTypes = []string{
			shared.TransactionTypeWriteOffReversal.Key(),
		}
	case shared.ScheduleTypeADFeeReductionReversals,
		shared.ScheduleTypeGeneralFeeReductionReversals,
		shared.ScheduleTypeMinimalFeeReductionReversals,
		shared.ScheduleTypeGAFeeReductionReversals,
		shared.ScheduleTypeGSFeeReductionReversals,
		shared.ScheduleTypeGTFeeReductionReversals:
		ledgerTypes = []string{
			shared.TransactionTypeFeeReductionReversal.Key(),
		}
		includeNegatives = true
	default:
		ledgerTypes = []string{
			shared.TransactionTypeUnknown.Key(),
		}
	}

	switch *a.ScheduleType {
	case shared.ScheduleTypeADFeeReductions,
		shared.ScheduleTypeADManualCredits,
		shared.ScheduleTypeADManualDebits,
		shared.ScheduleTypeADWriteOffs,
		shared.ScheduleTypeADWriteOffReversals,
		shared.ScheduleTypeADFeeReductionReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeAD.Key(),
		}
	case shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeGeneralWriteOffs,
		shared.ScheduleTypeGeneralWriteOffReversals,
		shared.ScheduleTypeGeneralFeeReductionReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeS2.Key(),
			shared.InvoiceTypeB2.Key(),
			shared.InvoiceTypeSF.Key(),
			shared.InvoiceTypeSE.Key(),
			shared.InvoiceTypeSO.Key(),
		}
	case shared.ScheduleTypeMinimalFeeReductions,
		shared.ScheduleTypeMinimalManualCredits,
		shared.ScheduleTypeMinimalManualDebits,
		shared.ScheduleTypeMinimalWriteOffs,
		shared.ScheduleTypeMinimalWriteOffReversals,
		shared.ScheduleTypeMinimalFeeReductionReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeS3.Key(),
			shared.InvoiceTypeB3.Key(),
			shared.InvoiceTypeSF.Key(),
			shared.InvoiceTypeSE.Key(),
			shared.InvoiceTypeSO.Key(),
		}
	case shared.ScheduleTypeGAFeeReductions,
		shared.ScheduleTypeGAManualCredits,
		shared.ScheduleTypeGAManualDebits,
		shared.ScheduleTypeGAWriteOffs,
		shared.ScheduleTypeGAWriteOffReversals,
		shared.ScheduleTypeGAFeeReductionReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeGA.Key(),
		}
	case shared.ScheduleTypeGSFeeReductions,
		shared.ScheduleTypeGSManualCredits,
		shared.ScheduleTypeGSManualDebits,
		shared.ScheduleTypeGSWriteOffs,
		shared.ScheduleTypeGSWriteOffReversals,
		shared.ScheduleTypeGSFeeReductionReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeGS.Key(),
		}
	case shared.ScheduleTypeGTFeeReductions,
		shared.ScheduleTypeGTManualCredits,
		shared.ScheduleTypeGTManualDebits,
		shared.ScheduleTypeGTWriteOffs,
		shared.ScheduleTypeGTWriteOffReversals,
		shared.ScheduleTypeGTFeeReductionReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeGT.Key(),
		}
	}

	return []any{a.Date.Time.Format("2006-01-02"), ledgerTypes, supervisionLevel, invoiceTypes, includeNegatives}
}
