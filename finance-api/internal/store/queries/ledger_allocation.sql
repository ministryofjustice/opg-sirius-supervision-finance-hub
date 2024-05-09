-- name: CreateLedgerAllocationForFeeReduction :one
insert into ledger_allocation (id, ledger_id, invoice_id, datetime, amount, status, reference,
                                                   notes, allocateddate, batchnumber, source,
                                                   transaction_type)
VALUES (nextval('ledger_allocation_id_seq'::regclass), $1, $2, now(), $3, 'Confirmed', null, null, null, null, null,
        null) returning *;
