-- name: CreatePendingCollection :exec
INSERT INTO pending_collection (id, finance_client_id, collection_date, amount, status, created_at, created_by)
VALUES (NEXTVAL('pending_collection_id_seq'),
        (SELECT id FROM finance_client WHERE client_id = @client_id),
        @collection_date,
        @amount,
        'PENDING',
        NOW(),
        @created_by);

-- name: GetPendingCollectionsForDate :many
SELECT pc.id, pc.amount, fc.court_ref
FROM pending_collection pc
         JOIN supervision_finance.finance_client fc ON fc.id = pc.finance_client_id
WHERE pc.collection_date = @date_collected::DATE
  AND pc.ledger_id IS NULL;

-- name: MarkPendingCollectionAsCollected :exec
UPDATE pending_collection
SET ledger_id = @ledger_id
WHERE id = @id;
