package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type NonReceiptTransactions struct {
	Date *shared.Date
}

const NonReceiptTransactionsQuery = `WITH transaction_totals AS (
   SELECT
        tt.line_description AS line_description,
        TO_CHAR(l.created_at, 'DD/MM/YYYY') AS transaction_date,
        tt.account_code AS account_code,
        (ABS(SUM(la.amount) / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS amount,
		cc.cost_centre,
		CASE WHEN tt.fee_type IN ('MCR','ZR','ZE','ZH','WO') THEN true ELSE false END AS is_credit
    FROM
        supervision_finance.ledger_allocation la
        JOIN supervision_finance.ledger l ON l.id = la.ledger_id
        JOIN supervision_finance.invoice i ON i.id = la.invoice_id
         LEFT JOIN LATERAL (
                        SELECT CASE WHEN i.feetype IN  ('AD', 'GA', 'GT', 'GS') THEN i.feetype ELSE (SELECT COALESCE(ifr.supervisionlevel, '')
                        FROM supervision_finance.invoice_fee_range ifr
                        WHERE ifr.invoice_id = i.id
                        ORDER BY id DESC
                        LIMIT 1) END AS supervision_level
                ) sl ON TRUE
                LEFT JOIN LATERAL (
                        SELECT CASE WHEN i.feetype IN ('GA', 'GS', 'GT') THEN '10486000' ELSE '10482009' END AS cost_centre LIMIT 1
    ) cc ON TRUE
        JOIN supervision_finance.transaction_type tt
                  ON l.type = tt.ledger_type AND sl.supervision_level = tt.supervision_level
    WHERE tt.is_receipt = false AND TO_CHAR(l.created_at, 'YYYY-MM-DD') = $1
    GROUP BY
        tt.line_description, TO_CHAR(l.created_at, 'DD/MM/YYYY'), tt.account_code, cc.cost_centre, tt.fee_type
	UNION
	SELECT tt.line_description AS line_description, TO_CHAR(i.created_at, 'DD/MM/YYYY') AS transaction_date, tt.account_code AS account_code, (ABS(SUM(i.amount) / 100.0)::NUMERIC(10, 2))::VARCHAR(255) AS amount, cc.cost_centre AS cost_centre, false AS is_credit 
	FROM supervision_finance.invoice i LEFT JOIN LATERAL (SELECT CASE WHEN i.feetype IN ('GA', 'GS', 'GT') THEN '10486000' ELSE '10482009' END AS cost_centre LIMIT 1) cc ON TRUE LEFT JOIN LATERAL (SELECT CASE WHEN i.feetype IN ('AD', 'GA', 'GT', 'GS') THEN i.feetype ELSE (SELECT COALESCE(ifr.supervisionlevel, '') FROM supervision_finance.invoice_fee_range ifr WHERE ifr.invoice_id = i.id ORDER BY id DESC LIMIT 1) END AS supervision_level) sl ON TRUE LEFT JOIN supervision_finance.transaction_type tt ON i.feetype = tt.fee_type AND sl.supervision_level = tt.supervision_level 
	WHERE TO_CHAR(i.created_at, 'YYYY-MM-DD') = $1 GROUP BY tt.line_description, TO_CHAR(i.created_at, 'DD/MM/YYYY'), tt.account_code, cc.cost_centre
),
partitioned_data AS (
    SELECT
        *,
        ROW_NUMBER() OVER (PARTITION BY line_description ORDER BY line_description) AS row_num
    FROM
        transaction_totals 
	CROSS JOIN (select 1 as n union all select 2) n
)
SELECT
    '0470' AS "Entity",
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
            '1816102003'
        END AS "Account",
    '0000000' AS "Objective",
    '00000000' AS "Analysis",
    '0000' AS "Intercompany",
    '00000000' AS "Spare",
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
    line_description || ' [' || transaction_date || ']' AS "Line description"
FROM
    partitioned_data
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
