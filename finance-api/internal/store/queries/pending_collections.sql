-- name: CreatePendingCollection :exec
INSERT INTO pending_collection (id, finance_client_id, collection_date, amount, created_at, created_by)
VALUES (NEXTVAL('pending_collection_id_seq'),
        (SELECT id FROM finance_client WHERE client_id = @client_id),
        @collection_date,
        @amount,
        NOW(),
        @created_by);
