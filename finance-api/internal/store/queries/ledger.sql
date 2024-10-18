-- name: CreateLedger :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, created_at, created_by, reference, method)
SELECT nextval('ledger_id_seq'),
       now(),
       fc.id,
       $2,
       $3,
       $4,
       $5,
       $6,
       now(),
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc WHERE client_id = $1
RETURNING id;

-- name: CreateLedgerForCaseRecNumber :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, created_at, created_by, reference, method)
SELECT nextval('ledger_id_seq'),
       $2,
       fc.id,
       $3,
       $4,
       $5,
       $6,
       now(),
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc WHERE caserecnumber = $1
RETURNING id;

-- name: GetLedgerForPayment :one
SELECT l.id
FROM ledger l
LEFT JOIN finance_client fc ON fc.id = l.finance_client_id
WHERE l.amount = $1 AND l.status = 'APPROVED' AND l.datetime = $2 AND l.type = $3 AND fc.caserecnumber = $4
LIMIT 1;