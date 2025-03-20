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
			ELSE 7
			END AS index
	FROM transaction_type WHERE is_receipt = TRUE
),
transaction_totals AS (
	SELECT 
		tt.line_description AS line_description,
		l.bankdate AS transaction_date, 
		CASE 
			WHEN l.type = 'SUPERVISION BACS PAYMENT' 
			THEN '1841102088' 
			ELSE '1841102050' END AS account_code,
		((SUM(ABS(la.amount)) / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS amount,
		n,
		tt.index
	FROM supervision_finance.ledger_allocation la 
	INNER JOIN supervision_finance.ledger l ON l.id = la.ledger_id 
	INNER JOIN LATERAL (
		SELECT tto.index, fee_type, line_description
		FROM transaction_type tt
		INNER JOIN transaction_type_order tto ON tt.id = tto.id
		WHERE tt.ledger_type = l.type
	) tt ON TRUE
	CROSS JOIN (SELECT 1 AS n UNION ALL SELECT 2) n
	WHERE l.created_at::DATE = $1
	GROUP BY tt.line_description, l.bankdate, l.type, n, tt.index
)
SELECT 	
	'="0470"'                                              		AS "Entity",
	'99999999'                                       			AS "Cost Centre",
	CASE WHEN n % 2 = 1 
		THEN account_code
		ELSE '1816102003'
		END                                             		AS "Account",
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
