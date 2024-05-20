-- name: CreateLedgerForFeeReduction :one
insert into ledger (id, reference, datetime, method, amount, notes, type, status, finance_client_id,
                                        parent_id, fee_reduction_id, confirmeddate, bankdate, batchnumber, bankaccount,
                                        line, source,
                                        createddate, createdby_id)
VALUES (nextval('ledger_id_seq'::regclass), gen_random_uuid(), now(), $1, $2, $3, $4, 'Status', $5, null, $6, null,
        null, null, null, null, null, now(), $7) returning *;

-- name: UpdateLedgerAdjustment :one
WITH filtered_ledger_allocation AS (
    SELECT lc.id
    from ledger l
             inner join ledger_allocation lc on lc.ledger_id = l.id
    where l.id = $1
      and l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF')
)
UPDATE ledger_allocation
SET status = 'APPROVED'
FROM filtered_ledger_allocation fla
WHERE ledger_allocation.id = fla.id
returning ledger_allocation.*;