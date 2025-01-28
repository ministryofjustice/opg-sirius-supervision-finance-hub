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
        SUM(la.amount) AS amount
    FROM
        supervision_finance.ledger_allocation la
        JOIN supervision_finance.ledger l ON l.id = la.ledger_id
        JOIN supervision_finance.invoice i ON i.id = la.invoice_id
        LEFT JOIN LATERAL ( SELECT ifr.supervisionlevel AS supervision_level
                            FROM supervision_finance.invoice_fee_range ifr
                            WHERE ifr.invoice_id = i.id
                            ORDER BY id
                            LIMIT 1 ) sl ON TRUE
        JOIN supervision_finance.transaction_type tt
                  ON l.type = tt.ledger_type AND sl.supervision_level = tt.supervision_level
    WHERE tt.is_receipt = false AND l.created_at = $1
    GROUP BY
        tt.line_description, TO_CHAR(l.created_at, 'DD/MM/YYYY'), tt.account_code
),
partitioned_data AS (
    SELECT
        *,
        ROW_NUMBER() OVER (PARTITION BY account_code ORDER BY account_code) AS row_num
    FROM
        transaction_totals
)
SELECT
    '0470' AS "Entity",
    CASE
        WHEN row_num % 2 = 1 THEN
            '10482009'
        ELSE
            '99999999'
        END AS "Cost Centre",
    CASE
        WHEN row_num % 2 = 1 THEN
            account_code
        ELSE
            '1816100000'
        END AS "Account",
    '0000000' AS "Objective",
    '00000000' AS "Analysis",
    '0000' AS "Intercompany",
    '00000000' AS "Spare",
    CASE
        WHEN row_num % 2 = 1 THEN
            ''
        ELSE
            amount::VARCHAR
        END AS "Debit",
    CASE
        WHEN row_num % 2 = 1 THEN
            amount::VARCHAR
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
        WHEN line_description LIKE 'AD Rem/Exem%' THEN 12
        WHEN line_description LIKE 'Gen Rem/Exem%' THEN 13
        WHEN line_description LIKE 'Min Rem/Exem%' THEN 14
        WHEN line_description LIKE 'AD Manual credit%' THEN 15
        WHEN line_description LIKE 'Gen Manual credit%' THEN 16
        WHEN line_description LIKE 'Min Manual credit%' THEN 17
        WHEN line_description LIKE 'AD Manual debit%' THEN 18
        WHEN line_description LIKE 'Gen Manual debit%' THEN 19
        WHEN line_description LIKE 'Min Manual debit%' THEN 20
        WHEN line_description LIKE 'AD Write-off%' THEN 21
        WHEN line_description LIKE 'Gen Write-off%' THEN 22
        WHEN line_description LIKE 'Min Write-off%' THEN 23
        WHEN line_description LIKE 'AD Write-off reversal%' THEN 24
        WHEN line_description LIKE 'Gen Write-off reversal%' THEN 25
        WHEN line_description LIKE 'Min Write-off reversal%' THEN 26
        ELSE 27
        END;`

func (n *NonReceiptTransactions) GetHeaders() []string {
	return []string{
		"Entity",
		"Cost Centre",
		"Account",
		"Objective",
		"Analysis",
		"Intercompany",
		"Spare",
		"Credit",
		"Line description",
		"Revenue cost centre",
		"Revenue cost centre description",
		"Revenue account code",
		"Revenue account code description",
		"Invoice type",
		"Trx number",
		"Transaction description",
		"Invoice date",
		"Due date",
		"Financial year",
		"Payment terms",
		"Original amount",
		"Outstanding amount",
		"Current",
		"0-1 years",
		"1-2 years",
		"2-3 years",
		"3-5 years",
		"5+ years",
		"Debt impairment years",
	}
}

func (n *NonReceiptTransactions) GetQuery() string {
	return NonReceiptTransactionsQuery
}

func (n *NonReceiptTransactions) GetParams() []any {
	return []any{n.Date.Time.Format("2006-01-02")}
}
