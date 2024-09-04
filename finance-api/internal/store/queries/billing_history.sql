SET SEARCH_PATH TO supervision_finance;

-- name: GetPendingLedgerAllocations :many
SELECT i.id invoice_id, l.id ledger_id, i.reference, l.type, la.amount, l.notes, l.confirmeddate, l.created_by, l.status, l.datetime
FROM ledger_allocation la
         JOIN ledger l ON l.id = la.ledger_id
         JOIN invoice i ON i.id = la.invoice_id
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
  AND l.status = 'PENDING'
ORDER BY l.datetime DESC;

-- name: GetGeneratedInvoices :many
SELECT i.id invoice_id, reference, feetype, amount, created_by, created_at
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
ORDER BY created_at DESC;

-- name: GetFeeReductionEvents :many
SELECT
   fr.type,
   fr.startdate,
   fr.enddate,
   fr.datereceived,
   fr.notes,
   fr.created_at,
   fr.created_by,
   fr.cancelled_at,
   fr.cancelled_by,
   fr.cancellation_reason,
   l.status,
   l.amount,
   l.datetime ledger_date,
   fc.client_id,
   i.id invoice_id,
   i.reference reference
FROM fee_reduction fr
JOIN finance_client fc ON fc.id = fr.finance_client_id
LEFT JOIN ledger l ON l.fee_reduction_id = fr.id
LEFT JOIN (SELECT DISTINCT ON (ledger_id) * FROM ledger_allocation) la ON l.id = la.ledger_id
LEFT JOIN invoice i ON i.id = la.invoice_id
WHERE fc.client_id = $1
AND (fr.created_at IS NOT NULL OR fr.cancelled_at IS NOT NULL);
