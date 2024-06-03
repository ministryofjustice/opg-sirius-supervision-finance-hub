SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, COALESCE(SUM(la.amount), 0)::int received
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status IN ('ALLOCATED', 'APPROVED')
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate
ORDER BY i.raiseddate DESC;

-- name: GetInvoiceBalance :one
SELECT i.amount initial, i.amount - COALESCE(SUM(la.amount), 0) outstanding, i.feetype
FROM invoice i
         LEFT JOIN ledger_allocation la on i.id = la.invoice_id
    AND la.status IN ('ALLOCATED', 'APPROVED')
WHERE i.id = $1
group by i.amount, i.feetype;

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