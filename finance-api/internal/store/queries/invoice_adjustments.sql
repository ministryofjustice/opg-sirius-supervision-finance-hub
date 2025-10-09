-- name: GetInvoiceAdjustments :many
SELECT ia.id,
       i.reference AS invoice_ref,
       ia.raised_date,
       ia.adjustment_type,
       ia.amount,
       ia.notes,
       ia.status,
       ia.created_by
FROM invoice_adjustment ia
         JOIN invoice i ON i.id = ia.invoice_id
         JOIN finance_client fc ON fc.id = ia.finance_client_id
WHERE fc.client_id = $1
ORDER BY ia.raised_date DESC, ia.created_at DESC;

-- name: CreatePendingInvoiceAdjustment :one
INSERT INTO invoice_adjustment (id, finance_client_id, invoice_id, raised_date, adjustment_type, amount, notes, status,
                                created_at, created_by)
SELECT NEXTVAL('invoice_adjustment_id_seq'),
       fc.id,
       $2,
       NOW(),
       $3,
       $4,
       $5,
       'PENDING',
       NOW(),
       $6
FROM finance_client fc
WHERE fc.client_id = $1
RETURNING (SELECT reference invoicereference FROM invoice WHERE id = invoice_id);

-- name: SetAdjustmentDecision :one
UPDATE invoice_adjustment ia
SET status     = $2,
    updated_at = NOW(),
    updated_by = $3
WHERE ia.id = $1
RETURNING ia.amount, ia.adjustment_type, ia.finance_client_id, ia.invoice_id,
    (SELECT (i.amount - COALESCE(SUM(la.amount), 0)) outstanding
     FROM invoice i
              LEFT JOIN ledger_allocation la ON i.id = la.invoice_id
              AND la.status NOT IN ('PENDING', 'UN ALLOCATED')
              AND la.ledger_id IN (SELECT id FROM ledger WHERE status = 'CONFIRMED')
     WHERE i.id = ia.invoice_id
     GROUP BY i.amount)::INT AS outstanding;

-- name: CreateLedgerForAdjustment :one
WITH created AS (
    INSERT INTO ledger (id, datetime, finance_client_id, amount, notes, type, status, fee_reduction_id, created_at,
                        created_by, reference, method)
        SELECT NEXTVAL('ledger_id_seq'),
               NOW(),
               fc.id,
               $2,
               $3,
               $4,
               $5,
               $6,
               NOW(),
               $7,
               gen_random_uuid(),
               ''
        FROM finance_client fc
        WHERE client_id = $1
        RETURNING id)
UPDATE invoice_adjustment ia
SET ledger_id = created.id
FROM created
WHERE ia.id = $8
RETURNING created.id;
