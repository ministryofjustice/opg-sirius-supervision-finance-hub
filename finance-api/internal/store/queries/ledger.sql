-- name: CreateLedger :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, createdby_id, reference, method)
SELECT nextval('ledger_id_seq'),
       now(),
       fc.id,
       $2,
       $3,
       $4,
       $5,
       $6,
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc WHERE client_id = $1
RETURNING id;

-- name: UpdateLedgerAdjustment :exec
UPDATE ledger l
SET status = $1
WHERE l.id = $2;

-- name: GetLedger :one
SELECT l.amount, l.notes, l.type, la.invoice_id
FROM ledger l
LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
WHERE l.id = $1;
