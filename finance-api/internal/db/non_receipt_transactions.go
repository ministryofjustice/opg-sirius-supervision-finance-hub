package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type NonReceiptTransactions struct {
	ReportQuery
	NonReceiptTransactionsInput
}

type NonReceiptTransactionsInput struct {
	Date *shared.Date
}

func NewNonReceiptTransactions(input NonReceiptTransactionsInput) ReportQuery {
	return &NonReceiptTransactions{
		ReportQuery:                 NewReportQuery(NonReceiptTransactionsQuery),
		NonReceiptTransactionsInput: input,
	}
}

const NonReceiptTransactionsQuery = `
WITH transaction_type_order AS (
	SELECT 
		id, 
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
			WHEN line_description LIKE 'AD Fee reduction reversal%' THEN 48
			WHEN line_description LIKE 'General Fee reduction reversal%' THEN 49
			WHEN line_description LIKE 'Minimal Fee reduction reversal%' THEN 50
			WHEN line_description LIKE 'GA Fee reduction reversal%' THEN 51
			WHEN line_description LIKE 'GS Fee reduction reversal%' THEN 52
			WHEN line_description LIKE 'GT Fee reduction reversal%' THEN 53
			ELSE 54
			END AS index
	FROM supervision_finance.transaction_type WHERE is_receipt = FALSE
),
transactions AS (
   	SELECT
		l.type AS ledger_type,
		i.feetype AS fee_type,
		la.amount AS amount,
		i.id AS invoice_id
	FROM
        supervision_finance.ledger_allocation la
        INNER JOIN supervision_finance.ledger l ON l.id = la.ledger_id
		INNER JOIN supervision_finance.invoice i ON i.id = la.invoice_id
	WHERE l.general_ledger_date = $1 AND la.status = 'ALLOCATED'
	UNION ALL
	SELECT
		NULL AS ledger_type,
		i.feetype AS fee_type,
		-i.amount AS amount,
		i.id AS invoice_id
	FROM supervision_finance.invoice i
	WHERE i.created_at::DATE = $1
),
transaction_totals AS (
	SELECT
		tt.line_description,
		tt.account_code,
		ABS(SUM(t.amount) / 100.0)::NUMERIC(10,2)::VARCHAR(255) AS amount,
		account.cost_centre,
		SUM(t.amount) >= 0 AS is_credit,
		tt.index,
		n
	FROM transactions t
	INNER JOIN LATERAL (
		SELECT CASE WHEN t.fee_type IN ('AD', 'GA', 'GS', 'GT') THEN t.fee_type ELSE (
			SELECT COALESCE(ifr.supervisionlevel, '')
			FROM supervision_finance.invoice_fee_range ifr
			WHERE ifr.invoice_id = t.invoice_id
			ORDER BY id DESC
			LIMIT 1) END AS supervision_level
	) sl ON TRUE
	INNER JOIN LATERAL (
		SELECT tto.index, fee_type, account_code, line_description 
		FROM supervision_finance.transaction_type tt
		INNER JOIN transaction_type_order tto ON tt.id = tto.id
		WHERE (tt.ledger_type = t.ledger_type OR (t.ledger_type IS NULL AND tt.fee_type = t.fee_type)) 
		AND sl.supervision_level = tt.supervision_level
	) tt ON TRUE
	INNER JOIN supervision_finance.account ON tt.account_code = account.code
	CROSS JOIN (SELECT 1 AS n UNION ALL SELECT 2) n
	GROUP BY tt.line_description, tt.account_code, account.cost_centre, tt.index, n 
)
SELECT
    '="0470"' AS "Entity",
    CASE WHEN n % 2 = 1 
		THEN cost_centre
        ELSE '99999999'
        END AS "Cost Centre",
    CASE WHEN n % 2 = 1 
		THEN account_code
        ELSE '1816102003'
        END AS "Account",
    '="0000000"' AS "Objective",
    '="00000000"' AS "Analysis",
    '="0000"' AS "Intercompany",
    '="00000000"' AS "Spare",
    CASE WHEN n % 2 = 1 AND is_credit = FALSE OR n % 2 = 0 AND is_credit 
		THEN ''
        ELSE amount
        END AS "Debit",
    CASE WHEN n % 2 = 1 AND is_credit = FALSE OR n % 2 = 0 AND is_credit 
		THEN amount
        ELSE ''
        END AS "Credit",
    line_description || ' [' || TO_CHAR($1, 'DD/MM/YYYY') || ']' AS "Line description"
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

func (n *NonReceiptTransactions) GetParams() []any {
	return []any{n.Date.Time.Format("2006-01-02")}
}
