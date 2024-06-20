-- name: GetInvoiceAdjustments :many
SELECT l.id,
       i.reference AS invoice_ref,
       l.datetime  AS raised_date,
       l.type,
       l.amount,
       l.notes,
       l.status
FROM supervision_finance.ledger l
         INNER JOIN supervision_finance.ledger_allocation lc ON lc.ledger_id = l.id
         INNER JOIN supervision_finance.invoice i ON i.id = lc.invoice_id
         INNER JOIN supervision_finance.finance_client fc ON fc.id = i.finance_client_id
WHERE fc.client_id = $1
  AND l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF', 'DEBIT MEMO')
ORDER BY l.datetime DESC;
