-- name: ChangePendingCollectionDate :execrows
UPDATE pending_collection
SET collection_date = '2026-05-26'
WHERE collection_date = '2026-05-25';
