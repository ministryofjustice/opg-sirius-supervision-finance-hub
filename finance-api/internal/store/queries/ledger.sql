-- name: CreateLedgerForFeeReduction :one
insert into ledger (id, reference, datetime, method, amount, notes, type, status, finance_client_id,
                                        parent_id, fee_reduction_id, confirmeddate, bankdate, batchnumber, bankaccount,
                                        line, source,
                                        createddate, createdby_id)
VALUES (nextval('ledger_id_seq'::regclass), gen_random_uuid(), now(), $1, $2, $3, $4, 'Status', $5, null, $6, null,
        null, null, null, null, null, now(), $7) returning *;

-- name: UpdateLedgerAdjustment :exec
UPDATE ledger l
SET status = $1
WHERE l.id = $2;