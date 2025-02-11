package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type ReceiptTransactions struct {
	Date *shared.Date
}

const ReceiptTransactionsQuery = `WITH transaction_totals AS (SELECT tt.line_description                 AS line_description,
                                  TO_CHAR(l.created_at, 'DD/MM/YYYY') AS transaction_date,
                                  tt.account_code                     AS account_code,
                                  ((SUM(la.amount) / 100.0)::NUMERIC(10, 2))::VARCHAR(255)                  AS amount,
								  l.type AS ledger_type
                           FROM supervision_finance.ledger_allocation la
                                    JOIN supervision_finance.ledger l ON l.id = la.ledger_id
                                    JOIN supervision_finance.invoice i ON i.id = la.invoice_id
                                    LEFT JOIN LATERAL (SELECT CASE WHEN i.feetype = 'AD' THEN 'AD' ELSE COALESCE(ifr.supervisionlevel, '') END AS supervision_level
                                                        FROM supervision_finance.invoice_fee_range ifr
                                                        WHERE ifr.invoice_id = i.id
                                                        ORDER BY id
                                                        LIMIT 1 ) sl ON TRUE
                                    JOIN supervision_finance.transaction_type tt
                                         ON l.type = tt.ledger_type AND (CASE
                                                                             WHEN l.type = 'BACS TRANSFER'
                                                                                 THEN l.bankaccount
                                                                             ELSE COALESCE(sl.supervision_level, '') END
                                                                            ) = tt.supervision_level
                           WHERE tt.is_receipt = true AND TO_CHAR(l.created_at, 'YYYY-MM-DD') = $1
                           GROUP BY tt.line_description, TO_CHAR(l.created_at, 'DD/MM/YYYY'), tt.account_code, l.type)
, partitioned_data AS (SELECT *,
                                ROW_NUMBER() OVER (PARTITION BY account_code ORDER BY account_code) AS row_num
                         FROM transaction_totals CROSS JOIN (select 1 as n union all select 2) n)
SELECT '0470'                                              AS "Entity",
	  '99999999'                                       AS "Cost Centre",
      CASE
          WHEN row_num % 2 = 1 THEN
              CASE WHEN ledger_type = 'SUPERVISION BACS PAYMENT' THEN '1841102088' ELSE '1841102050' END
          ELSE
              '1816100000'
          END                                             AS "Account",
      '0000000'                                           AS "Objective",
      '00000000'                                          AS "Analysis",
      '0000'                                              AS "Intercompany",
      CASE
          WHEN row_num % 2 = 1 THEN
              '000000'
          ELSE
              '00000'
          END                                          AS "Spare",
      CASE
          WHEN row_num % 2 = 1 THEN
              amount
          ELSE
              ''
          END                                             AS "Debit",
      CASE
          WHEN row_num % 2 = 1 THEN
              ''
          ELSE
              amount
          END                                             AS "Credit",
      line_description || ' [' || transaction_date || ']' AS "Line description"
FROM partitioned_data 
ORDER BY CASE
            WHEN line_description LIKE 'MOTO card%' THEN 1
            WHEN line_description LIKE 'Online card%' THEN 2
            WHEN line_description LIKE 'OPG BACS%' THEN 3
            WHEN line_description LIKE 'Supervision BACS%' THEN 4
            WHEN line_description LIKE 'Direct debit%' THEN 5
            WHEN line_description LIKE 'Cheque payment%' THEN 6
            WHEN line_description LIKE 'Bounced cheque%' THEN 7
            ELSE 8
            END;`

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
