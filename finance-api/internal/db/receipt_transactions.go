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
		CASE WHEN id = 42 THEN 1
		WHEN id IN (41, 48) THEN 2
		WHEN id = 44 THEN 3
		WHEN id = 43 THEN 4
		WHEN id = 40 THEN 5
		WHEN id = 45 THEN 6
		ELSE 7
		END AS index
	FROM transaction_type WHERE is_receipt = true
),
transaction_totals AS (
	SELECT 
		tt.line_description AS line_description,
		l.bankdate AS transaction_date, 
		tt.account_code AS account_code,
		((SUM(ABS(la.amount)) / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS amount,
		l.type AS ledger_type,
		n,
		tt.index
	FROM supervision_finance.ledger_allocation la 
	LEFT JOIN supervision_finance.ledger l ON l.id = la.ledger_id 
	LEFT JOIN LATERAL (
		SELECT tto.index, fee_type, account_code, line_description
		FROM transaction_type tt
		LEFT JOIN transaction_type_order tto ON tt.id = tto.id
		WHERE tt.ledger_type = l.type
		AND is_receipt = true
		AND l.type != 'CREDIT REAPPLY'
	) tt ON TRUE
	CROSS JOIN (select 1 AS n union all select 2) n
	WHERE l.created_at::DATE = $1
	GROUP BY tt.line_description, l.bankdate, tt.account_code, l.type, n, tt.index
)
SELECT 	
	'="0470"'                                              		AS "Entity",
	'99999999'                                       			AS "Cost Centre",
	CASE WHEN n % 2 = 1 
		THEN CASE WHEN ledger_type = 'SUPERVISION BACS PAYMENT' 
			THEN '1841102088' 
			ELSE '1841102050' 
			END
		ELSE '1816100000'
		END                                             			AS "Account",
	'="0000000"'                                           		AS "Objective",
	'="00000000"'                                          		AS "Analysis",
	'="0000"'                                              		AS "Intercompany",
	CASE WHEN n % 2 = 1 
		THEN '="000000"'
		ELSE '="00000"'
		END                                          					AS "Spare",
	CASE WHEN n % 2 = 1 
		THEN amount
		ELSE ''
		END                                             			AS "Debit",
	CASE WHEN n % 2 = 1 
		THEN ''
		ELSE amount
		END                                             			AS "Credit",
	line_description || ' [' || TO_CHAR(transaction_date, 'DD/MM/YYYY') || ']' AS "Line description"
FROM transaction_totals 
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
