-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, i.cacheddebtamount FROM invoice i inner join finance_client fc on fc.id = i.finance_client_id  where fc.client_id = $1 order by i.raiseddate desc;

-- name: GetLedgerAllocations :many
select la.id, la.amount, la.datetime, l.bankdate, l.type from ledger_allocation la inner join ledger l on la.ledger_id = l.id where la.invoice_id = $1 order by la.id desc;

-- name: GetSupervisionLevels :many
select supervisionlevel, fromdate, todate, amount from invoice_fee_range where invoice_id = $1 order by todate desc;
