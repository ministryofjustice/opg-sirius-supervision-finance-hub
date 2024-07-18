SET SEARCH_PATH TO supervision_finance;

-- name: GetPendingLedgerAllocations :many
SELECT i.id invoice_id, l.id ledger_id, i.reference, l.type, la.amount, l.notes, l.confirmeddate, l.createdby_id, l.status, l.datetime
FROM ledger_allocation la
         JOIN ledger l ON l.id = la.ledger_id
         JOIN invoice i ON i.id = la.invoice_id
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
  AND l.status = 'PENDING'
ORDER BY l.datetime DESC;

-- name: GetGeneratedInvoices :many
SELECT i.id invoice_id, reference, feetype, amount, createdby_id, coalesce(confirmeddate, createddate) invoice_date
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
ORDER BY COALESCE(confirmeddate, createddate) DESC;

-- name: GetFeeReductionEvents :many
SELECT fr.type, fr.startdate, fr.enddate, fr.datereceived, fr.notes, fr.created_at, fr.created_by, fr.cancelled_at, fr.cancelled_by, fr.cancellation_reason
FROM fee_reduction fr
JOIN finance_client fc ON fc.id = fr.finance_client_id
WHERE fc.client_id = $1
AND fr.created_at IS NOT NULL OR fr.cancelled_at IS NOT NULL;