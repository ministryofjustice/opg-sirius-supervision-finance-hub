-- name: SetPaymentMethod :exec
WITH this_payment_method AS (
    INSERT INTO payment_method (id, finance_client_id, type, created_by, created_at)
VALUES (NEXTVAL('payment_method_id_seq'),
        (SELECT id FROM finance_client WHERE client_id = @client_id),
        @payment_method,
        @created_by,
        now()))
UPDATE finance_client
 SET payment_method = @payment_method
 WHERE client_id = @client_id;
