-- name: CreateLedgerAllocation :exec
WITH this_ledger AS (SELECT id, datetime
                     FROM ledger l
                     WHERE l.id = @ledger_id::INT),
     allocation AS (INSERT INTO ledger_allocation (id, datetime, ledger_id, invoice_id, amount, status, notes)
         SELECT NEXTVAL('ledger_allocation_id_seq'),
                this_ledger.datetime,
                @ledger_id::INT,
                sqlc.narg('invoice_id')::INT,
                @amount::INT,
                @status::TEXT,
                @notes
         FROM this_ledger
         WHERE this_ledger.id = @ledger_id::INT)
UPDATE invoice i
SET cacheddebtamount = CASE
                           WHEN @status::TEXT = 'UNAPPLIED' THEN cacheddebtamount
                           ELSE COALESCE(cacheddebtamount, i.amount) - @amount::INT END
WHERE sqlc.narg('invoice_id')::INT IS NOT NULL AND i.id = sqlc.narg('invoice_id')::INT;
