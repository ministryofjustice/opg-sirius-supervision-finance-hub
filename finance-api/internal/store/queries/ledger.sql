-- name: CreateLedgerForFeeReduction :one
INSERT INTO supervision_finance.ledger (id, reference, datetime, method, amount, notes, type, status, finance_client_id,
                                        parent_id, fee_reduction_id, confirmeddate, bankdate, batchnumber, bankaccount,
                                        line, source,
                                        createddate, createdby_id)
VALUES (NEXTVAL('supervision_finance.ledger_id_seq'::REGCLASS), gen_random_uuid(), NOW(), $1, $2, $3, $4, 'Status', $5,
        NULL, $6, NULL,
        NULL, NULL, NULL, NULL, NULL, NOW(), $7)
RETURNING *;

-- name: UpdateLedgerAdjustment :exec
UPDATE supervision_finance.ledger l
SET status = $1
WHERE l.id = $2;

-- name: CreateLedger :one
WITH fc AS (SELECT id FROM supervision_finance.finance_client WHERE client_id = $1),
     ledger AS (
         INSERT INTO supervision_finance.ledger (id, datetime, amount, notes, type, finance_client_id, reference,
                                                 method, status)
             SELECT NEXTVAL('supervision_finance.ledger_id_seq'),
                    NOW(),
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
INTO supervision_finance.ledger_allocation (id, ledger_id, invoice_id, datetime, amount, status, notes)
SELECT NEXTVAL('supervision_finance.ledger_allocation_id_seq'),
       ledger.id,
       $2,
       ledger.datetime,
       $3,
       'PENDING',
       $4
FROM ledger
RETURNING (SELECT reference invoicereference FROM supervision_finance.invoice WHERE id = invoice_id);