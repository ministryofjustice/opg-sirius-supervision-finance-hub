-- name: CreateLedgerAllocation :exec
WITH this_ledger as (
    SELECT id, datetime FROM ledger WHERE id = $1
)
INSERT INTO ledger_allocation (id, datetime, ledger_id, invoice_id, amount, status, notes)
SELECT nextval('ledger_allocation_id_seq'),
       this_ledger.datetime,
       $1,
       $2,
       $3,
       $4,
       $5
FROM this_ledger WHERE this_ledger.id = $1;
