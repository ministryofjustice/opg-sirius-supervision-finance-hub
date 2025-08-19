package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type UnappliedTransactions struct {
	ReportQuery
	UnappliedTransactionsInput
}

type UnappliedTransactionsInput struct {
	Date *shared.Date
}

func NewUnappliedTransactions(input UnappliedTransactionsInput) ReportQuery {
	return &UnappliedTransactions{
		ReportQuery:                NewReportQuery(UnappliedTransactionsQuery),
		UnappliedTransactionsInput: input,
	}
}

const UnappliedTransactionsQuery = `
WITH transactions AS (
	SELECT
		SUM(CASE WHEN la.status = 'UNAPPLIED' AND i.id IS NOT NULL THEN ABS(la.amount) ELSE 0 END) AS unapplied_amount,
		SUM(CASE WHEN la.status = 'REAPPLIED' AND i.id IS NOT NULL THEN ABS(la.amount) ELSE 0 END) AS reapplied_amount,
		SUM(CASE WHEN la.status = 'REAPPLIED' AND i.id IS NULL THEN ABS(la.amount) ELSE 0 END) AS refund_amount,
		n
		FROM supervision_finance.ledger_allocation la 
		JOIN supervision_finance.ledger l ON l.id = la.ledger_id
		LEFT JOIN supervision_finance.invoice i ON la.invoice_id = i.id
		CROSS JOIN (SELECT 1 AS n UNION ALL SELECT 2) n
		WHERE l.general_ledger_date = $1
		AND la.status IN ('UNAPPLIED', 'REAPPLIED')
		GROUP BY n
		),
  splits AS (
	SELECT
	    CASE WHEN n % 2 = 1 
		THEN (refund_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)
		ELSE ''
		END AS debit,
	    CASE WHEN n % 2 = 1 
		THEN ''
		ELSE (refund_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)
		END AS credit,
	    'Bankline refund' AS line_description
	FROM transactions
	WHERE refund_amount > 0
	    UNION ALL 
	    SELECT
	    CASE WHEN n % 2 = 1 
		THEN (unapplied_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)
		ELSE ''
		END AS debit,
	    CASE WHEN n % 2 = 1 
		THEN ''
		ELSE (unapplied_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)
		END AS credit,
	    'Unapplied payments' AS line_description
	FROM transactions
	WHERE unapplied_amount > 0
	    UNION ALL 
	    SELECT
	    CASE WHEN n % 2 = 1 
		THEN (reapplied_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)
		ELSE ''
		END AS debit,
	    CASE WHEN n % 2 = 1 
		THEN ''
		ELSE (reapplied_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255)
		END AS credit,
	    'Reapplied payments' AS line_description
	FROM transactions
 	WHERE reapplied_amount > 0
	)
SELECT 	
	''                                              	AS "Entity",
	''                                       			AS "Cost Centre",
	''                                             		AS "Account",
	''                                           		AS "Objective",
	''                                          		AS "Analysis",
	''                                              	AS "Intercompany",
	''                                          		AS "Spare",
	debit                                             	AS "Debit",
	credit                                             	AS "Credit",
	line_description || ' [' || TO_CHAR($1, 'DD/MM/YYYY') || ']' AS "Line description"
FROM splits;
`

func (u *UnappliedTransactions) GetHeaders() []string {
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

func (u *UnappliedTransactions) GetParams() []any {
	return []any{u.Date.Time.Format("2006-01-02")}
}
