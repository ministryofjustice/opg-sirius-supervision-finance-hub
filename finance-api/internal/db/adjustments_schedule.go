package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type AdjustmentsSchedule struct {
	ReportQuery
	AdjustmentsScheduleParams
}

type AdjustmentsScheduleParams struct {
	Date         *shared.Date
	ScheduleType *shared.ScheduleType
}

func NewAdjustmentsSchedule(params AdjustmentsScheduleParams) ReportQuery {
	return &AdjustmentsSchedule{
		ReportQuery:               NewReportQuery(AdjustmentsScheduleQuery),
		AdjustmentsScheduleParams: params,
	}
}

const AdjustmentsScheduleQuery = `SELECT
	fc.court_ref AS "Court reference",
	i.reference AS "Invoice reference",
	(ABS(la.amount) / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS "Amount",
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
	)
	switch *a.ScheduleType {
	case shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeGeneralWriteOffs,
		shared.ScheduleTypeGeneralWriteOffReversals:
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeMinimalFeeReductions,
		shared.ScheduleTypeMinimalManualCredits,
		shared.ScheduleTypeMinimalManualDebits,
		shared.ScheduleTypeMinimalWriteOffs,
		shared.ScheduleTypeMinimalWriteOffReversals:
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
		shared.ScheduleTypeADWriteOffReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeAD.Key(),
		}
	case shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeGeneralWriteOffs,
		shared.ScheduleTypeGeneralWriteOffReversals:
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
		shared.ScheduleTypeMinimalWriteOffReversals:
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
		shared.ScheduleTypeGAWriteOffReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeGA.Key(),
		}
	case shared.ScheduleTypeGSFeeReductions,
		shared.ScheduleTypeGSManualCredits,
		shared.ScheduleTypeGSManualDebits,
		shared.ScheduleTypeGSWriteOffs,
		shared.ScheduleTypeGSWriteOffReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeGS.Key(),
		}
	case shared.ScheduleTypeGTFeeReductions,
		shared.ScheduleTypeGTManualCredits,
		shared.ScheduleTypeGTManualDebits,
		shared.ScheduleTypeGTWriteOffs,
		shared.ScheduleTypeGTWriteOffReversals:
		invoiceTypes = []string{
			shared.InvoiceTypeGT.Key(),
		}
	}

	return []any{a.Date.Time.Format("2006-01-02"), ledgerTypes, supervisionLevel, invoiceTypes}
}
