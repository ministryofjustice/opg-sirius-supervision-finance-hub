-- name: GetRefunds :many
SELECT r.id,
       r.raised_date,
       r.fulfilled_date,
       r.amount,
       r.status,
       r.notes,
       r.created_by,
       COALESCE(bd.name, '')::VARCHAR      AS account_name,
       COALESCE(bd.account, '')::VARCHAR   AS account_code,
       COALESCE(bd.sort_code, '')::VARCHAR AS sort_code
FROM refund r
         LEFT JOIN bank_details bd ON r.id = bd.refund_id
WHERE client_id = $1
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
    INSERT INTO refund (id, client_id, raised_date, amount, status, notes, created_by, created_at)
        VALUES (NEXTVAL('refund_id_seq'),
                $1,
                NOW(),
                $2,
                'PENDING',
                $3,
                $4,
                NOW())
        RETURNING id),
     b AS (
         INSERT INTO bank_details (id, refund_id, name, account, sort_code)
             SELECT NEXTVAL('refund_id_seq'), r.id, $5, $6, $7
             FROM r)
SELECT id
FROM r;
