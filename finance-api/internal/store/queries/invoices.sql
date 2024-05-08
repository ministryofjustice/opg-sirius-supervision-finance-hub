SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, i.cacheddebtamount
FROM invoice i
         inner join finance_client fc on i.finance_client_id = fc.id
where fc.client_id = $1 order by i.raiseddate desc;

-- name: GetLedgerAllocations :many
select la.id, la.amount, la.datetime, l.bankdate, l.type from ledger_allocation la inner join ledger l on la.ledger_id = l.id where la.invoice_id = $1 order by la.id desc;

-- name: GetSupervisionLevels :many
select supervisionlevel, fromdate, todate, amount from invoice_fee_range where invoice_id = $1 order by todate desc;

-- name: AddFeeReductionToInvoices :many
WITH filtered_invoices AS (
    SELECT i.id AS invoice_id, fr.id AS fee_reduction_id
    FROM invoice i
             JOIN fee_reduction fr
                  ON i.finance_client_id = fr.finance_client_id
    WHERE i.raiseddate >= (fr.datereceived - interval '6 months')
      AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
      AND fr.id = $1
)
UPDATE invoice i
SET fee_reduction_id = fi.fee_reduction_id
FROM filtered_invoices fi
WHERE i.id = fi.invoice_id
returning i.*;