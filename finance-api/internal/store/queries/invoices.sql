-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, COALESCE(SUM(la.amount), 0)::INT received
FROM supervision_finance.invoice i
         JOIN supervision_finance.finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN supervision_finance.ledger_allocation la
                   ON i.id = la.invoice_id AND la.status IN ('ALLOCATED', 'APPROVED')
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate
ORDER BY i.raiseddate DESC;

-- name: GetInvoiceBalance :one
SELECT i.amount initial, i.amount - COALESCE(SUM(la.amount), 0) outstanding, i.feetype
FROM supervision_finance.invoice i
         LEFT JOIN supervision_finance.ledger_allocation la on i.id = la.invoice_id
    AND la.status IN ('ALLOCATED', 'APPROVED')
WHERE i.id = $1
GROUP BY i.amount, i.feetype;

-- name: GetLedgerAllocations :many
SELECT la.id, la.amount, la.datetime, l.bankdate, l.type, la.status
FROM supervision_finance.ledger_allocation la
         INNER JOIN supervision_finance.ledger l ON la.ledger_id = l.id
WHERE la.invoice_id = $1
ORDER BY la.id DESC;

-- name: GetSupervisionLevels :many
SELECT supervisionlevel, fromdate, todate, amount
FROM supervision_finance.invoice_fee_range
WHERE invoice_id = $1
ORDER BY todate DESC;

-- name: AddFeeReductionToInvoices :many
WITH filtered_invoices AS (SELECT i.id AS invoice_id, fr.id AS fee_reduction_id
                           FROM supervision_finance.invoice i
                                    JOIN supervision_finance.fee_reduction fr
                                         ON i.finance_client_id = fr.finance_client_id
                           WHERE i.raiseddate >= (fr.datereceived - INTERVAL '6 months')
                             AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
                             AND fr.id = $1)
UPDATE supervision_finance.invoice i
SET fee_reduction_id = fi.fee_reduction_id
FROM filtered_invoices fi
WHERE i.id = fi.invoice_id
RETURNING i.*;