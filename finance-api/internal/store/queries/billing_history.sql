SET SEARCH_PATH TO supervision_finance;

-- name: GetPendingInvoiceAdjustments :many
SELECT ia.invoice_id, i.reference, ia.adjustment_type, ia.amount, ia.notes, ia.created_at, ia.created_by
FROM invoice_adjustment ia
         JOIN invoice i ON i.id = ia.invoice_id
         JOIN finance_client fc ON fc.id = ia.finance_client_id
WHERE fc.client_id = $1
ORDER BY ia.raised_date DESC;

-- name: GetRejectedInvoiceAdjustments :many
SELECT ia.invoice_id, i.reference, ia.adjustment_type, ia.amount, ia.notes, ia.updated_at, ia.updated_by
FROM invoice_adjustment ia
         JOIN invoice i ON i.id = ia.invoice_id
         JOIN finance_client fc ON fc.id = ia.finance_client_id
WHERE fc.client_id = $1
AND status = 'REJECTED'
ORDER BY ia.raised_date DESC;

-- name: GetGeneratedInvoices :many
SELECT i.id invoice_id, reference, feetype, amount, created_by, created_at
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
ORDER BY created_at DESC;

-- name: GetFeeReductionEvents :many
SELECT fr.type,
       fr.startdate,
       fr.enddate,
       fr.datereceived,
       fr.notes,
       fr.created_at,
       fr.created_by,
       fr.cancelled_at,
       fr.cancelled_by,
       fr.cancellation_reason
FROM fee_reduction fr
         JOIN finance_client fc ON fc.id = fr.finance_client_id
         LEFT JOIN ledger l ON l.fee_reduction_id = fr.id
         LEFT JOIN (SELECT DISTINCT ON (ledger_id) * FROM ledger_allocation) la ON l.id = la.ledger_id
         LEFT JOIN invoice i ON i.id = la.invoice_id
WHERE fc.client_id = $1
  AND (fr.created_at IS NOT NULL OR fr.cancelled_at IS NOT NULL);

-- name: GetLedgerAllocationsForClient :many
SELECT l.id AS ledger_id,
       la.invoice_id,
       i.reference,
       COALESCE(fr.type, l.type),
       la.status,
       l.amount  AS ledger_amount,
       la.amount AS allocation_amount,
       l.datetime AS created_at,
       l.created_by
FROM ledger_allocation la
         JOIN ledger l ON l.id = la.ledger_id
         JOIN finance_client fc ON fc.id = l.finance_client_id
         LEFT JOIN invoice i ON la.invoice_id = i.id
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.client_id = $1
  AND la.status NOT IN ('PENDING', 'UN ALLOCATED');

-- name: GetPendingRefundsForBillingHistory :many
SELECT r.raised_date,
       r.amount,
       r.decision,
       r.notes,
       r.created_by,
       r.created_at,
       r.decision_by,
       r.decision_at,
       r.processed_at,
       r.cancelled_at,
       r.fulfilled_at,
       r.cancelled_by
FROM refund r
        JOIN finance_client fc ON fc.id = r.finance_client_id
WHERE fc.client_id = $1
AND r.decision = 'PENDING'
ORDER BY r.created_at DESC;