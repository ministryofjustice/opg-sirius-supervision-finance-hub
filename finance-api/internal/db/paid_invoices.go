package db

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"time"
)

type PaidInvoices struct {
	ReportQuery
	FromDate   *shared.Date
	ToDate     *shared.Date
	GoLiveDate time.Time
}

func NewPaidInvoices(fromDate *shared.Date, toDate *shared.Date, goLiveDate time.Time) ReportQuery {
	return &PaidInvoices{
		ReportQuery: NewReportQuery(PaidInvoicesQuery),
		FromDate:    fromDate,
		ToDate:      toDate,
		GoLiveDate:  goLiveDate,
	}
}

const PaidInvoicesQuery = `
WITH paid_invoices AS (SELECT i.id
                       FROM supervision_finance.ledger_allocation la
                                JOIN supervision_finance.ledger l ON l.id = la.ledger_id
                                JOIN supervision_finance.invoice i ON i.id = la.invoice_id
                                LEFT JOIN supervision_finance.invoice_adjustment ia ON i.id = ia.invoice_id
                       WHERE la.status IN ('ALLOCATED', 'REAPPLIED')
                         AND l.type NOT IN ('CREDIT WRITE OFF', 'WRITE OFF REVERSAL')
                         AND l.datetime::DATE BETWEEN $1 AND $2
                       GROUP BY i.id, i.amount
                       HAVING i.amount <= COALESCE(SUM(la.amount), 0))
SELECT CONCAT(p.firstname, ' ', p.surname)                           AS "Customer Name",
       p.caserecnumber                                               AS "Customer number",
       fc.sop_number                                                 AS "SOP number",
       '="0470"'                                                     AS "Entity",
       COALESCE(cc.code::VARCHAR, 'Refer to SOP')                    AS "Revenue cost centre",
       COALESCE(cc.cost_centre_description::VARCHAR, 'Refer to SOP') AS "Revenue cost centre description",
       COALESCE(a.code::VARCHAR, 'Refer to SOP')                     AS "Revenue account code",
       COALESCE(a.account_code_description::VARCHAR, 'Refer to SOP') AS "Revenue account code description",
       i.feetype                                                     AS "Invoice type",
       i.reference                                                   AS "Invoice number",
       CASE
           WHEN tt.fee_type IS NULL THEN 'Refer to SOP'
           ELSE CONCAT(tt.fee_type, i.reference) END                                                           AS "Txn number",
       COALESCE(tt.description, l.type)                                                                        AS "Txn description",
       ((i.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255)                                                      AS "Original amount",
       COALESCE(TO_CHAR(l.bankdate, 'YYYY-MM-DD'), '')                                                         AS "Received date",
       TO_CHAR(l.datetime, 'YYYY-MM-DD')                                                                       AS "Sirius upload date",
       CASE
           WHEN l.type IN (
                           'BACS TRANSFER', 'CARD PAYMENT',
                           'DIRECT DEBIT PAYMENT', 'ONLINE CARD PAYMENT', 'MOTO CARD PAYMENT',
                           'SUPERVISION BACS PAYMENT', 'OPG BACS PAYMENT', 'SUPERVISION CHEQUE PAYMENT'
               )
               THEN ((la.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255)
           ELSE '0' END                                              AS "Cash amount",
       CASE
           WHEN l.type IN (
                           'UNKNOWN CREDIT',
                           'CREDIT MEMO', 'CREDIT EXEMPTION', 'CREDIT HARDSHIP', 'CREDIT REMISSION', 'CREDIT REAPPLY'
               )
               THEN ((la.amount / 100.0)::NUMERIC(10, 2)::VARCHAR(255))
           ELSE '0' END                                              AS "Credit amount",
       CASE
           WHEN l.type IN ('DEBIT MEMO', 'UNKNOWN DEBIT') THEN ((la.amount / 100.0)::NUMERIC(10, 2))::VARCHAR(255)
           ELSE '0' END                                              AS "Adjustment amount",
       COALESCE(ia.notes, fr.notes, '')                              AS "Memo line description"
FROM paid_invoices pa
         JOIN supervision_finance.invoice i ON pa.id = i.id
         JOIN supervision_finance.ledger_allocation la ON pa.id = la.invoice_id
         JOIN supervision_finance.ledger l ON la.ledger_id = l.id
         JOIN supervision_finance.finance_client fc ON fc.id = l.finance_client_id
         JOIN public.persons p ON fc.client_id = p.id
         LEFT JOIN supervision_finance.invoice_adjustment ia ON i.id = ia.invoice_id
         LEFT JOIN supervision_finance.fee_reduction fr ON fr.id = l.fee_reduction_id
         INNER JOIN LATERAL (
    SELECT CASE
               WHEN i.feetype IN ('AD', 'GA', 'GS', 'GT') THEN i.feetype
               ELSE (SELECT COALESCE(ifr.supervisionlevel, '')
                     FROM supervision_finance.invoice_fee_range ifr
                     WHERE ifr.invoice_id = i.id
                     ORDER BY id DESC
                     LIMIT 1) END AS supervision_level
    ) sl ON TRUE
         LEFT JOIN supervision_finance.transaction_type tt ON
    CASE
        WHEN l.type = 'BACS TRANSFER' THEN
            CASE WHEN l.bankaccount = 'BACS' THEN 'SUPERVISION BACS PAYMENT' ELSE 'OPG BACS PAYMENT' END
        ELSE l.type
        END = tt.ledger_type
        AND
    CASE
        WHEN l.type IN
             ('BACS TRANSFER', 'CARD PAYMENT', 'DIRECT DEBIT PAYMENT', 'ONLINE CARD PAYMENT', 'MOTO CARD PAYMENT',
              'SUPERVISION BACS PAYMENT', 'OPG BACS PAYMENT', 'SUPERVISION CHEQUE PAYMENT')
            THEN ''
        ELSE sl.supervision_level
        END = tt.supervision_level
         LEFT JOIN supervision_finance.account a ON tt.account_code = a.code
         LEFT JOIN supervision_finance.cost_centre cc ON cc.code = a.cost_centre
WHERE la.status IN ('ALLOCATED', 'REAPPLIED')
  AND l.type NOT IN ('CREDIT WRITE OFF', 'WRITE OFF REVERSAL')
ORDER BY l.datetime;
`

func (p *PaidInvoices) GetHeaders() []string {
	return []string{
		"Customer name",
		"Customer number",
		"SOP number",
		"Entity",
		"Cost centre",
		"Cost centre description",
		"Account code",
		"Account code description",
		"Invoice type",
		"Invoice number",
		"Txn number",
		"Txn description",
		"Original amount",
		"Received date",
		"Sirius upload date",
		"Cash amount",
		"Credit amount",
		"Adjustment amount",
		"Memo line description",
	}
}

func (p *PaidInvoices) GetParams() []any {
	var (
		from, to time.Time
	)

	if p.FromDate == nil {
		from = p.GoLiveDate
	} else {
		from = p.FromDate.Time
	}

	if p.ToDate == nil {
		to = time.Now()
	} else {
		to = p.ToDate.Time
	}

	return []any{from.Format("2006-01-02"), to.Format("2006-01-02")}
}
