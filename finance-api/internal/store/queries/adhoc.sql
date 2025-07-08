-- name: GetClientsWithCredit :many
SELECT fc.id AS finance_client_id
FROM finance_client fc
         LEFT JOIN ledger l ON fc.id = l.finance_client_id
         LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
WHERE la.status IN ('UNAPPLIED', 'REAPPLIED')
GROUP BY fc.id
HAVING ABS(SUM(la.amount)) > 0;