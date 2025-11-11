-- name: GetPaymentMethods :many
SELECT pm.id,
       finance_client_id,
       type,
       created_by,
       created_at
FROM payment_method pm
WHERE pm.finance_client_id = $1
ORDER BY created_at DESC;

-- name: AddPaymentMethod :one
INSERT INTO payment_method (id, finance_client_id, type, created_by, created_at)
VALUES (NEXTVAL('payment_method_id_seq'),
        (SELECT id FROM finance_client WHERE client_id = @client_id),
        @type,
        @created_by,
        now())
RETURNING *;

