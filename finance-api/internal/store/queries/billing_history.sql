-- name: GetGeneratedInvoices :many
SELECT i.id invoice_id, reference, feetype, amount, createddate, createdby_id
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
ORDER BY createddate DESC;

-- name: GetAppliedLedgerAllocations :many
SELECT i.id invoice_id, l.id ledger_id, i.reference, l.type, la.amount, l.notes, l.createddate, l.createdby_id
FROM ledger_allocation la
         JOIN ledger l ON l.id = la.ledger_id
         JOIN invoice i ON i.id = la.invoice_id
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
AND l.status IN ('APPROVED', 'CONFIRMED')
ORDER BY l.createddate DESC;
