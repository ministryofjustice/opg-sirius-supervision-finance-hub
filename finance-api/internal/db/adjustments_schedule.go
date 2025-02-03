package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type AdjustmentsSchedule struct {
	Date         *shared.Date
	ScheduleType *shared.ScheduleType
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
    ) sl ON $3 <> ''
	WHERE l.created_at::DATE = $1 AND l.type = ANY($2) AND COALESCE(sl.supervision_level, '') = $3;
`

func (c *AdjustmentsSchedule) GetHeaders() []string {
	return []string{
		"Court reference",
		"Invoice reference",
		"Amount",
		"Created date",
	}
}

func (c *AdjustmentsSchedule) GetQuery() string {
	return AdjustmentsScheduleQuery
}

func (c *AdjustmentsSchedule) GetParams() []any {
	var (
		ledgerTypes      []string
		supervisionLevel string
	)
	switch *c.ScheduleType {
	case shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeGeneralWriteOffs:
		supervisionLevel = "GENERAL"
	case shared.ScheduleTypeMinimalFeeReductions,
		shared.ScheduleTypeMinimalManualCredits,
		shared.ScheduleTypeMinimalManualDebits,
		shared.ScheduleTypeMinimalWriteOffs:
		supervisionLevel = "MINIMAL"
	default:
		supervisionLevel = ""
	}

	switch *c.ScheduleType {
	case shared.ScheduleTypeADFeeReductions,
		shared.ScheduleTypeGeneralFeeReductions,
		shared.ScheduleTypeMinimalFeeReductions:
		ledgerTypes = []string{
			"CREDIT " + shared.TransactionTypeHardship.Key(),
			"CREDIT " + shared.TransactionTypeExemption.Key(),
			"CREDIT " + shared.TransactionTypeRemission.Key(),
		}
	case shared.ScheduleTypeADManualCredits,
		shared.ScheduleTypeGeneralManualCredits,
		shared.ScheduleTypeMinimalManualCredits:
		ledgerTypes = []string{
			shared.TransactionTypeCreditMemo.Key(),
		}
	case shared.ScheduleTypeADManualDebits,
		shared.ScheduleTypeGeneralManualDebits,
		shared.ScheduleTypeMinimalManualDebits:
		ledgerTypes = []string{
			shared.TransactionTypeDebitMemo.Key(),
		}
	case shared.ScheduleTypeADWriteOffs,
		shared.ScheduleTypeGeneralWriteOffs,
		shared.ScheduleTypeMinimalWriteOffs:
		ledgerTypes = []string{
			shared.TransactionTypeWriteOff.Key(),
		}
	default:
		ledgerTypes = []string{
			shared.TransactionTypeUnknown.Key(),
		}
	}

	return []any{c.Date.Time.Format("2006-01-02 15:04:05"), ledgerTypes, supervisionLevel}
}
