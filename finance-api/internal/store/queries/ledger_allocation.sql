-- name: CreateLedgerAllocationForFeeReduction :one
insert into ledger_allocation (id, ledger_id, invoice_id, datetime, amount, status, reference,
                                                   notes, allocateddate, batchnumber, source,
                                                   transaction_type)
VALUES (nextval('ledger_allocation_id_seq'::regclass), $1, $2, now(), $3, 'Confirmed', null, null, null, null, null,
        null) returning *;

-- name: UpdateLedgerAllocationAdjustment :exec
WITH filtered_ledger_allocation AS (
    SELECT lc.id
    from ledger l
             inner join ledger_allocation lc on lc.ledger_id = l.id
    where l.id = $1
)
UPDATE ledger_allocation
SET status = 'APPROVED'
FROM filtered_ledger_allocation fla
WHERE ledger_allocation.id = fla.id;