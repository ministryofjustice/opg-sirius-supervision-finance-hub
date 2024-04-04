-- name: GetInvoices :many
SELECT id, reference, amount, raiseddate, cacheddebtamount FROM invoice WHERE finance_client_id = $1 order by raiseddate desc;

-- name: GetLedgerAllocations :many
select id, amount, allocateddate, status from ledger_allocation where invoice_id = $1;
