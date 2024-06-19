-- name: GetAccountInformation :one
WITH balances AS (SELECT i.id, fc.client_id, i.amount, COALESCE(SUM(la.amount), 0) paid
                  FROM supervision_finance.finance_client fc
                           LEFT JOIN supervision_finance.invoice i ON fc.id = i.finance_client_id
                           LEFT JOIN supervision_finance.ledger_allocation la
                                     ON i.id = la.invoice_id AND la.status IN ('ALLOCATED', 'APPROVED')
                  WHERE fc.client_id = $1
                  GROUP BY i.id, fc.client_id)
SELECT COALESCE(SUM(balances.amount), 0) - SUM(balances.paid) outstanding, 0 credit, fc.payment_method
FROM supervision_finance.finance_client fc
         JOIN balances ON fc.client_id = balances.client_id
WHERE fc.client_id = $1
GROUP BY fc.payment_method;
