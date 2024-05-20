-- name: CreateLedgerAllocationForFeeReduction :one
insert into ledger_allocation (id, ledger_id, invoice_id, datetime, amount, status, reference,
                                                   notes, allocateddate, batchnumber, source,
                                                   transaction_type)
VALUES (nextval('ledger_allocation_id_seq'::regclass), $1, $2, now(), $3, 'Confirmed', null, null, null, null, null,
        null) returning *;

-- -- name: UpdateLedgerAllocationAdjustment :one
-- WITH filtered_ledger_allocation AS (
--     SELECT lc.id
--     from ledger l
--              inner join ledger_allocation lc on lc.ledger_id = l.id
--              inner join invoice i on i.id = lc.invoice_id
--              inner join finance_client fc on fc.id = i.finance_client_id
--     where fc.client_id = $1 and l.id = $2
--       and l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF')
-- )
-- UPDATE ledger_allocation
-- SET status = 'APPROVED'
-- FROM filtered_ledger_allocation fla
-- WHERE ledger_allocation.id = fla.id
-- returning ledger_allocation.*;

-- name: UpdateLedgerAllocationAdjustment :one
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