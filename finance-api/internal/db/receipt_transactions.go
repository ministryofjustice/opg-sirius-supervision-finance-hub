package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type ReceiptTransactions struct {
	Date *shared.Date
}

const ReceiptTransactionsQuery = `
WITH transaction_type_order AS (
    SELECT
        id,
        CASE
            WHEN line_description LIKE 'MOTO card%' THEN 1
            WHEN line_description LIKE 'Online card%' THEN 2
            WHEN line_description LIKE 'OPG BACS%' THEN 3
            WHEN line_description LIKE 'Supervision BACS%' THEN 4
            WHEN line_description LIKE 'Direct debit%' THEN 5
            WHEN line_description LIKE 'Cheque payment%' THEN 6
        END AS index
    FROM transaction_type WHERE is_receipt = TRUE
),
transaction_totals AS (
    SELECT
        tt.line_description || ' [' || TO_CHAR(l.bankdate, 'DD/MM/YYYY') || ']' AS line_description,
        CASE
            WHEN l.type = 'SUPERVISION BACS PAYMENT' THEN '1841102088'
            ELSE '1841102050'
        END AS account_code,
        SUM(ABS(la.amount)) AS debit_amount,
        SUM(CASE WHEN la.status != 'UNAPPLIED' THEN ABS(la.amount) ELSE 0 END) AS credit_amount,
        SUM(CASE WHEN la.status = 'UNAPPLIED' THEN ABS(la.amount) ELSE 0 END) AS unapply_amount,
        tt.index
    FROM supervision_finance.ledger_allocation la
    INNER JOIN supervision_finance.ledger l ON l.id = la.ledger_id
    INNER JOIN LATERAL (
        SELECT tto.index, fee_type, line_description
        FROM transaction_type tt
        INNER JOIN transaction_type_order tto ON tt.id = tto.id
        WHERE tt.ledger_type = l.type AND tto.index IS NOT NULL
    ) tt ON TRUE
    WHERE l.created_at::DATE = $1
    GROUP BY tt.line_description, l.bankdate, l.type, tt.index
),
transaction_rows AS (
    SELECT
        '="0470"' AS entity,
        '99999999' AS cost_centre,
        account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="000000"' AS spare,
        (debit_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS debit,
        '' AS credit,
        line_description,
        index,
        1 AS n
    FROM transaction_totals
    UNION ALL
    SELECT
        '="0470"' AS entity,
        '99999999' AS cost_centre,
        '1816100000' AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="00000"' AS spare,
        '' AS debit,
        (credit_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS credit,
        line_description,
        index,
        2 AS n
    FROM transaction_totals
    UNION ALL
    SELECT
        '="0470"' AS entity,
        '' AS cost_centre,
        '' AS account_code,
        '' AS objective,
        '' AS analysis,
        '' AS intercompany,
        '' AS spare,
        '' AS debit,
        (unapply_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS credit,
        line_description,
        index,
        3 AS n
    FROM transaction_totals
    WHERE unapply_amount > 0
)
SELECT
    entity AS "Entity",
    cost_centre AS "Cost Centre",
    account_code AS "Account",
    objective AS "Objective",
    analysis AS "Analysis",
    intercompany AS "Intercompany",
    spare AS "Spare",
    debit AS "Debit",
    credit AS "Credit",
    line_description AS "Line description"
FROM transaction_rows
ORDER BY index, n;`

func (r *ReceiptTransactions) GetHeaders() []string {
	return []string{
		"Entity",
		"Cost Centre",
		"Account",
		"Objective",
		"Analysis",
		"Intercompany",
		"Spare",
		"Debit",
		"Credit",
		"Line description",
	}
}

func (r *ReceiptTransactions) GetQuery() string {
	return ReceiptTransactionsQuery
}

func (r *ReceiptTransactions) GetParams() []any {
	return []any{r.Date.Time.Format("2006-01-02")}
}
