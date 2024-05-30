SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate
FROM invoice i
         INNER JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
ORDER BY i.raiseddate DESC;

-- name: GetLedgerAllocations :many
SELECT la.invoice_id, la.id, la.amount, la.datetime, l.bankdate, l.type, la.status
FROM ledger_allocation la
         INNER JOIN ledger l ON la.ledger_id = l.id
WHERE la.invoice_id = ANY($1::int[])
ORDER BY la.id DESC;

-- name: GetSupervisionLevels :many
SELECT invoice_id, supervisionlevel, fromdate, todate, amount
FROM invoice_fee_range
WHERE invoice_id = ANY($1::int[])
ORDER BY todate DESC;

-- name: GetInvoiceBalance :one
SELECT i.amount initial, i.amount - COALESCE(SUM(la.amount), 0) outstanding
FROM invoice i
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id
    AND la.status <> 'PENDING'
WHERE i.id = $1
GROUP BY i.amount;

-- name: AddFeeReductionToInvoices :many
WITH filtered_invoices AS (SELECT i.id AS invoice_id, fr.id AS fee_reduction_id
                           FROM invoice i
                                    JOIN fee_reduction fr
                                         ON i.finance_client_id = fr.finance_client_id
                           WHERE i.raiseddate >= (fr.datereceived - INTERVAL '6 months')
                             AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
                             AND fr.id = $1)
UPDATE invoice i
SET fee_reduction_id = fi.fee_reduction_id
FROM filtered_invoices fi
WHERE i.id = fi.invoice_id
RETURNING i.*;