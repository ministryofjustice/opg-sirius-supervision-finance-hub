package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type ReceiptTransactions struct {
	ReportQuery
	ReceiptTransactionsInput
}

type ReceiptTransactionsInput struct {
	Date *shared.Date
}

func NewReceiptTransactions(input ReceiptTransactionsInput) ReportQuery {
	return &ReceiptTransactions{
		ReportQuery:              NewReportQuery(ReceiptTransactionsQuery),
		ReceiptTransactionsInput: input,
	}
}

const ReceiptTransactionsQuery = `WITH 
    transaction_type_order AS (
    SELECT
        id,
        CASE
            WHEN line_description LIKE 'MOTO card%' THEN 1
            WHEN line_description LIKE 'Online card%' THEN 2
            WHEN line_description LIKE 'OPG BACS%' THEN 3
            WHEN line_description LIKE 'Supervision BACS%' THEN 4
            WHEN line_description LIKE 'Direct debit%' THEN 5
            WHEN line_description LIKE 'Cheque payment%' THEN 6
            WHEN line_description LIKE 'Refund%' THEN 7
        END AS index
	FROM supervision_finance.transaction_type
),
ledger_totals AS (
    SELECT
		CASE 
		    WHEN l.type = 'SUPERVISION CHEQUE PAYMENT' THEN tt.line_description || ' [' || l.pis_number || ']'
		    ELSE tt.line_description || ' [' || TO_CHAR(l.bankdate, 'DD/MM/YYYY') || ']' 
		END AS line_description,
        SUM(CASE WHEN l.amount > 0 THEN l.amount ELSE 0 END) AS debit_amount,
        SUM(CASE WHEN l.amount < 0 THEN ABS(l.amount) ELSE 0 END) AS reversal_amount,
        tt.index
    FROM supervision_finance.ledger l
    INNER JOIN LATERAL (
        SELECT tto.index, fee_type, line_description
        FROM transaction_type tt
        INNER JOIN transaction_type_order tto ON tt.id = tto.id
        WHERE tt.ledger_type = l.type AND tto.index IS NOT NULL
    ) tt ON TRUE
    WHERE l.created_at::DATE = $1
    GROUP BY tt.line_description, tt.index, l.type, l.pis_number, l.bankdate
),
allocation_totals AS (
	SELECT 
		CASE 
		    WHEN l.type = 'SUPERVISION CHEQUE PAYMENT' THEN tt.line_description || ' [' || l.pis_number || ']'
		    ELSE tt.line_description || ' [' || TO_CHAR(l.bankdate, 'DD/MM/YYYY') || ']' 
		    END AS line_description,
        CASE
            WHEN l.type = 'SUPERVISION BACS PAYMENT' THEN '1841102088'
            ELSE '1841102050'
        END AS debit_account_code,
		CASE 
		    WHEN l.type = 'REFUND' THEN '1816102005'
			ELSE '1816102003'
		END AS credit_account_code,
        SUM(CASE WHEN l.amount > 0 AND la.status != 'UNAPPLIED' AND la.amount > 0 THEN la.amount ELSE 0 END) AS credit_amount,
        SUM(CASE WHEN l.amount > 0 AND la.status = 'UNAPPLIED' AND la.amount < 0 THEN ABS(la.amount) ELSE 0 END) AS overpayment_amount,
        SUM(CASE WHEN l.amount < 0 AND la.status != 'UNAPPLIED' AND la.amount < 0 THEN ABS(la.amount) ELSE 0 END) AS reversed_credit_amount,
        SUM(CASE WHEN l.amount < 0 AND la.status = 'UNAPPLIED' THEN ABS(la.amount) ELSE 0 END) AS reversed_overpayment_amount,
        l.bankdate,
        l.pis_number,
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
	GROUP BY tt.line_description, tt.index, l.bankdate, l.type, l.pis_number
),
transaction_rows AS (
    -- debit row
    SELECT
        debit_account_code AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="000000"' AS spare,
        (lt.debit_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS debit,
        '' AS credit,
        at.line_description,
        at.bankdate,
        at.pis_number,
        at.index,
        1 AS n
    FROM allocation_totals at
    JOIN ledger_totals lt ON at.index = lt.index AND at.line_description = lt.line_description
    WHERE lt.debit_amount > 0
    UNION ALL
    -- credit row
    SELECT
		credit_account_code AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="00000"' AS spare,
        '' AS debit,
        (credit_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS credit,
        line_description,
        bankdate,
        pis_number,
        index,
        2 AS n
    FROM allocation_totals
    WHERE credit_amount > 0
    UNION ALL
        -- overpayment row
    SELECT
        '1816102005' AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="00000"' AS spare,
        '' AS debit,
        (overpayment_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS credit,
        line_description,
        bankdate,
        pis_number,
        index,
        3 AS n
    FROM allocation_totals
    WHERE overpayment_amount > 0
    UNION ALL
    -- reversal debit row
    SELECT
        at.debit_account_code AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="000000"' AS spare,
        '' AS debit,
        (lt.reversal_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS credit,
        at.line_description,
        at.bankdate,
        at.pis_number,
        at.index,
        4 AS n
    FROM allocation_totals at
    JOIN ledger_totals lt ON at.index = lt.index AND at.line_description = lt.line_description
    WHERE lt.reversal_amount > 0
    UNION ALL
    -- reversal credit row
    SELECT
		credit_account_code AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="00000"' AS spare,
        (reversed_credit_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS debit,
        '' AS credit,
        line_description,
        bankdate,
        pis_number,
        index,
        5 AS n
    FROM allocation_totals
    WHERE reversed_credit_amount > 0
    UNION ALL
    -- reversal overpayment row
    SELECT
        '1816102005' AS account_code,
        '="0000000"' AS objective,
        '="00000000"' AS analysis,
        '="0000"' AS intercompany,
        '="00000"' AS spare,
        (reversed_overpayment_amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255) AS debit,
        '' AS credit,
        line_description,
        bankdate,
        pis_number,
        index,
        6 AS n
    FROM allocation_totals
    WHERE reversed_overpayment_amount > 0
)
SELECT
    '="0470"' AS "Entity",
    '99999999' AS "Cost Centre",
    account_code AS "Account",
    objective AS "Objective",
    analysis AS "Analysis",
    intercompany AS "Intercompany",
    spare AS "Spare",
    debit AS "Debit",
    credit AS "Credit",
    line_description AS "Line description"
FROM transaction_rows
ORDER BY index, bankdate, pis_number, n;
`

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

func (r *ReceiptTransactions) GetParams() []any {
	return []any{r.Date.Time.Format("2006-01-02")}
}
