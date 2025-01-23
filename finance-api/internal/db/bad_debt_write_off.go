package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"os"
	"time"
)

type BadDebtWriteOff struct {
	FromDate *shared.Date
	ToDate   *shared.Date
}

const BadDebtWriteOffQuery = `SELECT CONCAT(p.firstname, ' ', p.surname)        "Customer Name",
       p.caserecnumber                            "Customer number",
       fc.sop_number                              "SOP number",
       '="0470"'                                  AS "Entity",
       cc.code                                 AS "Cost centre",
       ac.code                                  AS "Account code",
       ac.account_code_description              AS "Account code description",
       ((la.amount / 100.0)::NUMERIC(10, 2))::varchar(255)       AS "Adjustment amount",
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
  AND l.datetime BETWEEN $1 AND $2;`

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

func (b *BadDebtWriteOff) GetQuery() string {
	return BadDebtWriteOffQuery
}

func (b *BadDebtWriteOff) GetParams() []any {
	if b.FromDate == nil {
		from := shared.NewDate(os.Getenv("FINANCE_HUB_LIVE_DATE"))
		b.FromDate = &from
	}

	if b.ToDate == nil {
		to := shared.Date{Time: time.Now()}
		b.ToDate = &to
	}

	b.ToDate.Time = b.ToDate.Time.Truncate(24 * time.Hour).Add(24 * time.Hour)

	return []any{b.FromDate.Time.Format("2006-01-02 15:04:05"), b.ToDate.Time.Format("2006-01-02 15:04:05")}
}
