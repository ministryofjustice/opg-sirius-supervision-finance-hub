package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type NonReceiptTransactions struct {
	Date *shared.Date
}

const NonReceiptTransactionsQuery = `
WITH transactions AS (
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
		SELECT fee_type, account_code, line_description, is_receipt, CASE WHEN fee_type IN ('MCR', 'ZR', 'ZE', 'ZH', 'WO') THEN true ELSE false END AS is_credit
		FROM transaction_type tt
		WHERE (tt.ledger_type = t.ledger_type OR tt.fee_type = t.fee_type)
		AND sl.supervision_level = tt.supervision_level
		ORDER BY id DESC
	) tt ON TRUE
	LEFT JOIN account ON tt.account_code = account.code
	CROSS JOIN (select 1 as n union all select 2) n
	WHERE tt.is_receipt = false
	GROUP BY tt.line_description, t.created_at, tt.account_code, account.cost_centre, tt.is_credit, n
)
SELECT
    '="0470"' AS "Entity",
    CASE
        WHEN n % 2 = 1 THEN
            cost_centre
        ELSE
            '99999999'
        END AS "Cost Centre",
    CASE
        WHEN n % 2 = 1 THEN
            account_code
        ELSE
            '1816100000'
        END AS "Account",
    '="0000000"' AS "Objective",
    '="00000000"' AS "Analysis",
    '="0000"' AS "Intercompany",
    '="00000000"' AS "Spare",
    CASE
        WHEN n % 2 = 1 AND is_credit = false OR n % 2 = 0 AND is_credit THEN
            ''
        ELSE
            amount
        END AS "Debit",
    CASE
        WHEN n % 2 = 1 AND is_credit = false OR n % 2 = 0 AND is_credit THEN
            amount
        ELSE
            ''
        END AS "Credit",
    line_description || ' [' || TO_CHAR(transaction_date, 'DD/MM/YYYY') || ']' AS "Line description"
FROM
    transaction_totals
ORDER BY
    CASE
        WHEN line_description LIKE 'AD invoice%' THEN 1
        WHEN line_description LIKE 'S2 invoice%' THEN 2
        WHEN line_description LIKE 'S3 invoice%' THEN 3
        WHEN line_description LIKE 'B2 invoice%' THEN 4
        WHEN line_description LIKE 'B3 invoice%' THEN 5
        WHEN line_description LIKE 'Gen SF invoice%' THEN 6
        WHEN line_description LIKE 'Min SF invoice%' THEN 7
        WHEN line_description LIKE 'Gen SE invoice%' THEN 8
        WHEN line_description LIKE 'Min SE invoice%' THEN 9
        WHEN line_description LIKE 'Gen SO invoice%' THEN 10
        WHEN line_description LIKE 'Min SO invoice%' THEN 11
		WHEN line_description LIKE 'GA invoice%' THEN 12
		WHEN line_description LIKE 'GS invoice%' THEN 13
		WHEN line_description LIKE 'GT invoice%' THEN 14
        WHEN line_description LIKE 'AD Rem/Exem%' THEN 15
        WHEN line_description LIKE 'Gen Rem/Exem%' THEN 16
        WHEN line_description LIKE 'Min Rem/Exem%' THEN 17
		WHEN line_description LIKE 'GA remissions & hardships%' THEN 18
		WHEN line_description LIKE 'GS remissions & hardships%' THEN 19
		WHEN line_description LIKE 'GT remissions & hardships%' THEN 20
		WHEN line_description LIKE 'GA exemptions%' THEN 21
		WHEN line_description LIKE 'GS exemptions%' THEN 22
		WHEN line_description LIKE 'GT exemptions%' THEN 23
        WHEN line_description LIKE 'AD Manual credit%' THEN 24
        WHEN line_description LIKE 'Gen Manual credit%' THEN 25
        WHEN line_description LIKE 'Min Manual credit%' THEN 26
        WHEN line_description LIKE 'GA Manual credit%' THEN 27
        WHEN line_description LIKE 'GS Manual credit%' THEN 28
        WHEN line_description LIKE 'GT Manual credit%' THEN 29
        WHEN line_description LIKE 'AD Manual debit%' THEN 30
        WHEN line_description LIKE 'Gen Manual debit%' THEN 31
        WHEN line_description LIKE 'Min Manual debit%' THEN 32
        WHEN line_description LIKE 'GA Manual debit%' THEN 33
        WHEN line_description LIKE 'GS Manual debit%' THEN 34
        WHEN line_description LIKE 'GT Manual debit%' THEN 35
        WHEN line_description LIKE 'AD Write-off%' AND line_description NOT LIKE 'AD Write-off reversal%' THEN 36
        WHEN line_description LIKE 'Gen Write-off%' AND line_description NOT LIKE 'Gen Write-off reversal%' THEN 37
        WHEN line_description LIKE 'Min Write-off%' AND line_description NOT LIKE 'Min Write-off reversal%' THEN 38
        WHEN line_description LIKE 'GA Write-off%' AND line_description NOT LIKE 'GA Write-off reversal%' THEN 39
        WHEN line_description LIKE 'GS Write-off%' AND line_description NOT LIKE 'GS Write-off reversal%' THEN 40
        WHEN line_description LIKE 'GT Write-off%' AND line_description NOT LIKE 'GT Write-off reversal%' THEN 41
        WHEN line_description LIKE 'AD Write-off reversal%' THEN 42
        WHEN line_description LIKE 'Gen Write-off reversal%' THEN 43
        WHEN line_description LIKE 'Min Write-off reversal%' THEN 44
        WHEN line_description LIKE 'GA Write-off reversal%' THEN 45
        WHEN line_description LIKE 'GS Write-off reversal%' THEN 46
        WHEN line_description LIKE 'GT Write-off reversal%' THEN 47
        ELSE 48
        END, n;
`

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
