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
