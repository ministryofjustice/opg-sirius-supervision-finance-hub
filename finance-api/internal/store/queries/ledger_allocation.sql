-- name: CreateLedgerAllocationForFeeReduction :one
INSERT INTO ledger_allocation (id, ledger_id, invoice_id, datetime, amount, status, reference,
                               notes, allocateddate, batchnumber, source,
                               transaction_type)
VALUES (NEXTVAL('ledger_allocation_id_seq'::REGCLASS), $1, $2, NOW(), $3, 'Confirmed', NULL, NULL, NULL, NULL, NULL,
        NULL)
RETURNING *;

-- name: UpdateLedgerAllocationAdjustment :exec
UPDATE ledger_allocation la
SET status = $1
FROM ledger l
WHERE l.id = $2
  AND l.id = la.ledger_id
  AND l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF', 'DEBIT MEMO');