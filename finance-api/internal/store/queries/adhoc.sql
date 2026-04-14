-- name: PurgePendingCollections :execrows
DELETE FROM pending_collection WHERE status <> 'COLLECTED';