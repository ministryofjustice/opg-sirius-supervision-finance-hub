package db

import (
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type BadDebtWriteOff struct {
	ReportQuery
	BadDebtWriteOffInput
}

type BadDebtWriteOffInput struct {
	FromDate   *shared.Date
	ToDate     *shared.Date
	GoLiveDate time.Time
}

func NewBadDebtWriteOff(input BadDebtWriteOffInput) ReportQuery {
	return &BadDebtWriteOff{
		ReportQuery:          NewReportQuery(BadDebtWriteOffQuery),
		BadDebtWriteOffInput: input,
	}
}

const BadDebtWriteOffQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)        "Customer Name",
       p.caserecnumber                            "Customer number",
       fc.sop_number                              "SOP number",
       '="0470"'                                  AS "Entity",
       cc.code                                 AS "Cost centre",
       ac.code                                  AS "Account code",
       ac.account_code_description              AS "Account code description",
       ((la.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255)       AS "Adjustment amount",
       l.datetime                              AS "Adjustment date",
       CASE
           WHEN l.type = 'CREDIT WRITE OFF' THEN CONCAT('WO', i.reference)
           ELSE CONCAT('WOR', i.reference) END AS "Txn number",
       CONCAT(a.name, ' ', a.surname)          AS "Approver"
FROM supervision_finance.finance_client fc
         JOIN public.persons p ON fc.client_id = p.id
         JOIN supervision_finance.ledger l ON fc.id = l.finance_client_id
         JOIN supervision_finance.ledger_allocation la
              ON l.id = la.ledger_id AND la.status = 'ALLOCATED' -- to not include unapply/reapply
         JOIN supervision_finance.invoice i ON la.invoice_id = i.id
         LEFT JOIN public.assignees a ON l.created_by = a.id
         LEFT JOIN LATERAL (
    SELECT CASE WHEN i.feetype = 'AD' THEN 'AD' ELSE COALESCE(ifr.supervisionlevel, '') END AS supervision_level
    FROM supervision_finance.invoice_fee_range ifr
    WHERE ifr.invoice_id = i.id
    ORDER BY id DESC
    LIMIT 1
    ) sl ON TRUE
         JOIN supervision_finance.transaction_type tt
              ON CASE WHEN l.type = 'CREDIT WRITE OFF' THEN 'WO' ELSE 'WOR' END = tt.fee_type
			  AND CASE WHEN i.feetype = 'AD' THEN 'AD' ELSE sl.supervision_level END = tt.supervision_level
         JOIN supervision_finance.account ac ON tt.account_code = ac.code
         JOIN supervision_finance.cost_centre cc ON cc.code = ac.cost_centre
WHERE l.type IN ('CREDIT WRITE OFF', 'WRITE OFF REVERSAL')
  AND l.datetime::DATE BETWEEN $1 AND $2;`

func (b *BadDebtWriteOff) GetHeaders() []string {
	return []string{
		"Customer name",
		"Customer number",
		"SOP number",
		"Entity",
		"Cost centre",
		"Account code",
		"Account code description",
		"Adjustment amount",
		"Adjustment date",
		"Txn number",
		"Approver",
	}
}

func (b *BadDebtWriteOff) GetParams() []any {
	var (
		from, to time.Time
	)

	if b.FromDate == nil || b.FromDate.IsNull() {
		from = b.GoLiveDate
	} else {
		from = b.FromDate.Time
	}

	if b.ToDate == nil || b.ToDate.IsNull() {
		to = time.Now()
	} else {
		to = b.ToDate.Time
	}

	return []any{from.Format("2006-01-02"), to.Format("2006-01-02")}
}
