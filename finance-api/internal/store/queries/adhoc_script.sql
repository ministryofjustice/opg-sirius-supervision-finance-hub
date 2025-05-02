SET SEARCH_PATH TO supervision_finance;

-- name: GetNegativeInvoices :many
WITH total_debt AS (select i.id, fc.court_ref, i.reference, i.amount as invoiceamount, SUM(la.amount) as laamount
                    from invoice  i
                             inner join ledger_allocation la on la.invoice_id = i.id
                             inner join ledger l on l.id = la.ledger_id
                             inner join finance_client fc on i.finance_client_id = fc.id
                    where la.status NOT IN ('PENDING', 'UNALLOCATED') and l.status = 'CONFIRMED'
                    group by i.id, i.reference, fc.court_ref)
select td.id, td.reference, td.court_ref, sum(invoiceamount - laamount) as ledgerallocationamountneeded, i.person_id
from total_debt td
         inner join invoice i on i.id = td.id
group by td.id, td.reference, td.court_ref, i.person_id
HAVING sum(invoiceamount - laamount) < 0;
