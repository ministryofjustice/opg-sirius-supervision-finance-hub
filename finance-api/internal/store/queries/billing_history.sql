SET SEARCH_PATH TO supervision_finance;

-- name: GetClientLedgerAllocations :many
SELECT i.id invoice_id,
       l.id ledger_id,
       i.reference,
       l.type ledger_type,
       COALESCE((SELECT type FROM fee_reduction WHERE id = l.fee_reduction_id), '') fee_reduction_type,
       la.amount,
       l.notes,
       l.confirmeddate,
       l.createdby_id,
       l.status,
       l.datetime
FROM ledger_allocation la
         JOIN ledger l ON l.id = la.ledger_id
         JOIN invoice i ON i.id = la.invoice_id
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
  AND l.status IN ('PENDING', 'APPROVED')
ORDER BY l.datetime DESC;

-- name: GetGeneratedInvoices :many
SELECT i.id invoice_id, reference, feetype, amount, createdby_id, coalesce(confirmeddate, createddate) invoice_date
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
ORDER BY COALESCE(confirmeddate, createddate) DESC;
