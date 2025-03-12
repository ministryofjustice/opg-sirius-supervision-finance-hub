package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type NonReceiptTransactions struct {
	Date *shared.Date
}

const NonReceiptTransactionsQuery = `
WITH transaction_type_order AS (
	SELECT 
		id, 
		CASE WHEN id = 1 THEN 1 
		WHEN id = 2 THEN 2
		WHEN id = 3 THEN 3
		WHEN id = 4 THEN 4
		WHEN id = 5 THEN 5
		WHEN id = 6 THEN 6
		WHEN id = 7 THEN 7
		WHEN id = 8 THEN 8
		WHEN id = 9 THEN 9
		WHEN id = 1 THEN 10
		WHEN id = 11 THEN 11
		WHEN id = 33 THEN 12
		WHEN id = 34 THEN 13
		WHEN id = 35 THEN 14
		WHEN id IN (12, 13, 14) THEN 15
		WHEN id IN (19, 20, 21) THEN 16
		WHEN id IN (26, 27, 28) THEN 17
		WHEN id IN (49, 52) THEN 18
		WHEN id IN (50, 53) THEN 19
		WHEN id IN (51, 54) THEN 20
		WHEN id = 55 THEN 21
		WHEN id = 56 THEN 22
		WHEN id = 57 THEN 23
		WHEN id = 15 THEN 24
		WHEN id = 22 THEN 25
		WHEN id = 29 THEN 26
		WHEN id = 58 THEN 27
		WHEN id = 59 THEN 28
		WHEN id = 60 THEN 29
		WHEN id = 16 THEN 30
		WHEN id = 23 THEN 31
		WHEN id = 30 THEN 32
		WHEN id = 61 THEN 33
		WHEN id = 62 THEN 34
		WHEN id = 63 THEN 35
		WHEN id = 17 THEN 36
		WHEN id = 24 THEN 37
		WHEN id = 31 THEN 38
		WHEN id = 64 THEN 39
		WHEN id = 65 THEN 40
		WHEN id = 66 THEN 41
		WHEN id = 18 THEN 42
		WHEN id = 25 THEN 43
		WHEN id = 32 THEN 44
		WHEN id = 67 THEN 45
		WHEN id = 68 THEN 46
		WHEN id = 69 THEN 47
		ELSE 48
		END AS index
	FROM transaction_type WHERE is_receipt = false
),
transactions AS (
   	SELECT
		l.created_at::DATE AS created_at,
		l.type AS ledger_type,
		null AS fee_type,
		la.amount AS amount,
		la.invoice_id AS invoice_id
	FROM
        supervision_finance.ledger_allocation la
        LEFT JOIN ledger l ON l.id = la.ledger_id
	WHERE l.created_at::DATE = $1
	UNION
	SELECT
		i.created_at::DATE AS created_at,
		null AS ledger_type,
		i.feetype AS fee_type,
		i.amount AS amount,
		i.id AS invoice_id
	FROM supervision_finance.invoice i
	WHERE i.created_at::DATE = $1
),
transaction_totals AS (
	SELECT
		tt.line_description,
		t.created_at AS transaction_date,
		tt.account_code,
		ABS(SUM(t.amount) / 100.0)::NUMERIC(10,2)::VARCHAR(255) AS amount,
		account.cost_centre,
		tt.is_credit,
		tt.index,
		n
	FROM transactions t
	LEFT JOIN LATERAL (
		SELECT COALESCE(ifr.supervisionlevel, '') AS supervision_level
		FROM invoice_fee_range ifr
		WHERE ifr.invoice_id = t.invoice_id
		ORDER BY id DESC
		LIMIT 1
	) sl ON TRUE
	LEFT JOIN LATERAL (
		SELECT 
			tto.index, fee_type, account_code, line_description, 
			CASE WHEN fee_type IN ('MCR', 'ZR', 'ZE', 'ZH', 'WO') THEN true ELSE false END AS is_credit
		FROM transaction_type tt
		LEFT JOIN transaction_type_order tto ON tt.id = tto.id
		WHERE (tt.ledger_type = t.ledger_type OR tt.fee_type = t.fee_type) 
		AND is_receipt = false
		AND sl.supervision_level = tt.supervision_level
	) tt ON TRUE
	LEFT JOIN account ON tt.account_code = account.code
	CROSS JOIN (select 1 as n union all select 2) n
	GROUP BY tt.line_description, t.created_at, tt.account_code, account.cost_centre, tt.is_credit, tt.index, n 
)
SELECT
    '="0470"' AS "Entity",
    CASE WHEN n % 2 = 1 
		THEN cost_centre
        ELSE '99999999'
        END AS "Cost Centre",
    CASE WHEN n % 2 = 1 
		THEN account_code
        ELSE '1816100000'
        END AS "Account",
    '="0000000"' AS "Objective",
    '="00000000"' AS "Analysis",
    '="0000"' AS "Intercompany",
    '="00000000"' AS "Spare",
    CASE WHEN n % 2 = 1 AND is_credit = false OR n % 2 = 0 AND is_credit 
		THEN ''
        ELSE amount
        END AS "Debit",
    CASE WHEN n % 2 = 1 AND is_credit = false OR n % 2 = 0 AND is_credit 
		THEN amount
        ELSE ''
        END AS "Credit",
    line_description || ' [' || TO_CHAR(transaction_date, 'DD/MM/YYYY') || ']' AS "Line description"
FROM transaction_totals
ORDER BY index, n;`

func (n *NonReceiptTransactions) GetHeaders() []string {
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

func (n *NonReceiptTransactions) GetQuery() string {
	return NonReceiptTransactionsQuery
}

func (n *NonReceiptTransactions) GetParams() []any {
	return []any{n.Date.Time.Format("2006-01-02")}
}
