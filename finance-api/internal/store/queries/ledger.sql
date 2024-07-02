-- name: CreateLedgerForFeeReduction :one
insert into ledger (id, reference, datetime, method, amount, notes, type, status, finance_client_id,
                    parent_id, fee_reduction_id, confirmeddate, bankdate, batchnumber, bankaccount,
                    line, source, createddate, createdby_id)
VALUES (nextval('ledger_id_seq'::regclass), gen_random_uuid(), now(), $1, $2, $3, $4, 'APPROVED', $5, null, $6, null,
        null, null, null, null, null, now(), $7) returning *;

-- name: UpdateLedgerAdjustment :exec
UPDATE ledger l
SET status = $1
WHERE l.id = $2;

-- name: CreateLedger :one
WITH fc AS (SELECT id FROM finance_client WHERE client_id = $1),
     ledger AS (
         INSERT INTO ledger (id, datetime, amount, notes, type, finance_client_id, reference, method, status)
             SELECT nextval('ledger_id_seq'),
                    now(),
                    $3,
                    $4,
                    $5,
                    fc.id,
                    gen_random_uuid(),
                    '',
                    'PENDING'
             FROM fc
             RETURNING id, datetime)
INSERT
INTO ledger_allocation (id, ledger_id, invoice_id, datetime, amount, status, notes)
SELECT nextval('ledger_allocation_id_seq'),
       ledger.id,
       $2,
       ledger.datetime,
       $3,
       'PENDING',
       $4
FROM ledger
returning (SELECT reference invoiceReference FROM invoice WHERE id = invoice_id);