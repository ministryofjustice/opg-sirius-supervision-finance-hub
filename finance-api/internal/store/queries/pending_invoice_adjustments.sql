-- name: GetInvoiceAdjustments :many
select l.id,
       i.reference as invoice_ref,
       l.datetime as raised_date,
       l.type,
       l.amount,
       l.notes,
       l.status
from ledger l
         inner join ledger_allocation lc on lc.ledger_id = l.id
         inner join invoice i on i.id = lc.invoice_id
         inner join finance_client fc on fc.id = i.finance_client_id
where fc.client_id = $1
and l.type IN ('CREDIT MEMO', 'CREDIT WRITE OFF')
order by l.datetime desc;