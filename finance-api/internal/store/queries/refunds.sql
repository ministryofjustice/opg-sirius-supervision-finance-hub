-- name: GetRefunds :many
SELECT r.id,
       r.raised_date,
       r.fulfilled_at::DATE AS fulfilled_date,
       r.amount,
       CASE
           WHEN r.fulfilled_at IS NOT NULL THEN 'FULFILLED'
           WHEN r.cancelled_at IS NOT NULL THEN 'CANCELLED'
           WHEN r.processed_at IS NOT NULL THEN 'PROCESSING'
           ELSE r.decision
           END::VARCHAR      AS status,
       r.notes,
       r.created_by,
       COALESCE(bd.name, '')::VARCHAR      AS account_name,
       COALESCE(bd.account, '')::VARCHAR   AS account_code,
       COALESCE(bd.sort_code, '')::VARCHAR AS sort_code
FROM refund r
         JOIN finance_client fc ON fc.id = r.finance_client_id
         LEFT JOIN bank_details bd ON r.id = bd.refund_id
WHERE fc.client_id = $1
ORDER BY r.raised_date DESC, r.created_at DESC;

-- name: GetRefundAmount :one
SELECT ABS(COALESCE(SUM(
                            CASE
                                WHEN la.status IN ('UNAPPLIED', 'REAPPLIED')
                                    THEN la.amount
                                ELSE 0
                                END), 0))::INT AS credit
FROM finance_client fc
         LEFT JOIN ledger l ON fc.id = l.finance_client_id AND l.status = 'CONFIRMED'
         LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
WHERE fc.client_id = $1;

-- name: CreateRefund :one
WITH r AS (
    INSERT INTO refund (id, finance_client_id, raised_date, amount, decision, notes, created_by, created_at)
        VALUES (NEXTVAL('refund_id_seq'),
                (SELECT id FROM finance_client WHERE client_id = @client_id),
                NOW(),
                @amount,
                'PENDING',
                @notes,
                @created_by,
                NOW())
        RETURNING id),
     b AS (
         INSERT INTO bank_details (id, refund_id, name, account, sort_code)
             SELECT NEXTVAL('refund_id_seq'), r.id, @account_name, @account_number, @sort_code
             FROM r)
SELECT id
FROM r;

-- name: SetRefundDecision :exec
UPDATE refund
SET decision    = @decision,
    decision_at = NOW(),
    decision_by = @decision_by
WHERE finance_client_id = (SELECT id FROM finance_client WHERE client_id = @client_id)
  AND id = @refund_id;

-- name: RemoveBankDetails :exec
DELETE
FROM bank_details
WHERE refund_id = $1;


-- name: MarkRefundsAsProcessed :many
UPDATE refund
SET processed_at = NOW()
WHERE decision = 'APPROVED' AND processed_at IS NULL
RETURNING id;
