-- name: CreatePendingCollection :exec
INSERT INTO pending_collection (id, finance_client_id, collection_date, amount, status, created_at, created_by)
VALUES (NEXTVAL('pending_collection_id_seq'),
        (SELECT id FROM finance_client WHERE client_id = @client_id),
        @collection_date,
        @amount,
        'PENDING',
        NOW(),
        @created_by);

-- name: GetPendingOutstandingBalance :one
WITH finance_client_id AS (SELECT id
                           FROM finance_client
                           WHERE client_id = $1
                           LIMIT 1),
     debt AS (SELECT fc.id, SUM(i.amount) AS debt
              FROM finance_client_id fc
                       LEFT JOIN invoice i ON fc.id = i.finance_client_id
              GROUP BY fc.id),
     pending AS (SELECT fc.id, SUM(pc.amount) AS pending
                 FROM pending_collection pc
                          JOIN finance_client_id fc ON pc.finance_client_id = fc.id
                 WHERE status = 'PENDING'
                 GROUP BY fc.id),
     credit AS (SELECT fc.id, SUM(la.amount) AS credit
                FROM ledger l
                         JOIN ledger_allocation la ON l.id = la.ledger_id
                         JOIN finance_client_id fc ON l.finance_client_id = fc.id
                WHERE l.status = 'CONFIRMED'
                  AND (
                    la.status = 'ALLOCATED' OR
                    (la.status IN ('UNAPPLIED', 'REAPPLIED') AND la.invoice_id IS NOT NULL)
                    )
                GROUP BY fc.id)
SELECT (COALESCE(d.debt, 0) - COALESCE(c.credit, 0) - COALESCE(p.pending, 0))::INT
FROM debt d
         LEFT JOIN credit c ON c.id = d.id
         LEFT JOIN pending p ON p.id = d.id;

-- name: GetPendingCollectionsForDate :many
SELECT pc.id, pc.amount, fc.court_ref
FROM pending_collection pc
         JOIN supervision_finance.finance_client fc ON fc.id = pc.finance_client_id
WHERE pc.collection_date = @date_collected::DATE
  AND pc.status = 'PENDING';

-- name: CheckPendingCollection :one
SELECT pc.id
FROM pending_collection pc
JOIN supervision_finance.finance_client fc ON fc.id = pc.finance_client_id
WHERE pc.collection_date = @date_collected::DATE
    AND pc.amount = @amount
    AND fc.client_id = @client_id
    AND pc.status = 'PENDING';

-- name: MarkPendingCollectionAsCollected :exec
UPDATE pending_collection
SET ledger_id = @ledger_id,
    status    = 'COLLECTED'
WHERE id = @id;

-- name: GetPendingCollections :many
SELECT pc.id, pc.amount, pc.collection_date
FROM pending_collection pc
         JOIN finance_client fc ON pc.finance_client_id = fc.id
WHERE pc.status = 'PENDING'
  AND fc.client_id = @client_id
ORDER BY pc.collection_date;

-- name: CancelPendingCollection :exec
UPDATE pending_collection
SET status = 'CANCELLED'
WHERE id = @id;
