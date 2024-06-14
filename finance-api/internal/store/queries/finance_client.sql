-- name: GetAccountInformation :one
WITH balances AS (SELECT fc.client_id, COALESCE(SUM(i.amount), 0) total, COALESCE(SUM(la.amount), 0) paid
FROM finance_client fc
         LEFT JOIN invoice i ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status <> 'PENDING'
WHERE fc.client_id = $1
GROUP BY fc.client_id, i.amount)
SELECT SUM(balances.total) - SUM(balances.paid) outstanding, 0 credit, fc.payment_method
FROM finance_client fc
JOIN balances ON fc.client_id = balances.client_id
WHERE fc.client_id = $1
GROUP BY fc.payment_method;
