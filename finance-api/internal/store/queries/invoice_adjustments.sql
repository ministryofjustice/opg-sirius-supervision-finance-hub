-- name: GetInvoiceAdjustments :many
SELECT ia.id,
       i.reference AS invoice_ref,
       ia.raised_date,
       ia.adjustment_type,
       ia.amount,
       ia.notes,
       ia.status
FROM invoice_adjustments ia
         JOIN invoice i ON i.id = ia.invoice_id
         JOIN finance_client fc ON fc.id = ia.client_id
WHERE fc.client_id = $1
ORDER BY ia.raised_date DESC, ia.created_at DESC;

-- name: CreatePendingInvoiceAdjustment :one
INSERT INTO invoice_adjustments (id, client_id, invoice_id, raised_date, adjustment_type, amount, notes, status,
                                 created_at, created_by)
SELECT NEXTVAL('invoice_adjustments_id_seq'),
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

-- name: GetAdjustmentForDecision :one
SELECT ia.amount,
       ia.adjustment_type,
       ia.client_id,
       ia.invoice_id,
       i.amount - COALESCE(SUM(la.amount), 0) outstanding
FROM invoice_adjustments ia
         JOIN invoice i ON ia.invoice_id = i.id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
WHERE ia.id = $1
GROUP BY ia.amount, ia.adjustment_type, ia.client_id, ia.invoice_id, i.amount;

-- name: SetAdjustmentDecision :one
UPDATE invoice_adjustments ia
SET status     = $2,
    updated_at = NOW(),
    updated_by = $3
WHERE ia.id = $1
RETURNING ia.amount, ia.adjustment_type, ia.client_id, ia.invoice_id,
    (SELECT i.amount - COALESCE(SUM(la.amount), 0) outstanding
     FROM invoice i
              LEFT JOIN ledger_allocation la
                        ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
     WHERE i.id = ia.invoice_id
     GROUP BY i.amount);
