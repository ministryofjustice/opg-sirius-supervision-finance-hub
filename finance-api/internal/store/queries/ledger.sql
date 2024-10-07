-- name: CreateLedger :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, created_by, reference, method)
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

-- name: CreateLedgerForCaseRecNumber :one
INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, created_by, reference, method)
SELECT nextval('ledger_id_seq'),
       $2,
       fc.id,
       $3,
       $4,
       $5,
       $6,
       $7,
       gen_random_uuid(),
       ''
FROM finance_client fc WHERE caserecnumber = $1
RETURNING id;
