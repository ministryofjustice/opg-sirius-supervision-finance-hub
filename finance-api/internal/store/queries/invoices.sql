SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, i.cacheddebtamount
FROM invoice i
         inner join finance_client fc on fc.id = i.finance_client_id
where fc.client_id = $1
order by i.raiseddate desc;

-- name: GetInvoiceBalance :one
SELECT i.amount initial, i.amount - COALESCE(SUM(la.amount), 0) outstanding, i.feetype
FROM invoice i
         LEFT JOIN ledger_allocation la on i.id = la.invoice_id
    AND la.status <> 'PENDING'
WHERE i.id = $1
group by i.amount, i.feetype;

-- name: GetLedgerAllocations :many
select la.id, la.amount, la.datetime, l.bankdate, l.type, la.status
from ledger_allocation la
         inner join ledger l on la.ledger_id = l.id
where la.invoice_id = $1
order by la.id desc;

-- name: GetSupervisionLevels :many
select supervisionlevel, fromdate, todate, amount
from invoice_fee_range
where invoice_id = $1
order by todate desc;

-- name: AddFeeReductionToInvoices :many
WITH filtered_invoices AS (SELECT i.id AS invoice_id, fr.id AS fee_reduction_id
                           FROM invoice i
                                    JOIN fee_reduction fr
                                         ON i.finance_client_id = fr.finance_client_id
                           WHERE i.raiseddate >= (fr.datereceived - interval '6 months')
                             AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
                             AND fr.id = $1)
UPDATE invoice i
SET fee_reduction_id = fi.fee_reduction_id
FROM filtered_invoices fi
WHERE i.id = fi.invoice_id
returning i.*;

-- name: AddFeeReductionToInvoice :one
WITH filtered_invoices AS (SELECT i.id AS invoice_id, fr.id AS fee_reduction_id, fr.type, fr.finance_client_id
                           FROM invoice i
                                    JOIN fee_reduction fr
                                         ON i.finance_client_id = fr.finance_client_id
                           WHERE i.raiseddate >= (fr.datereceived - interval '6 months')
                             AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
                             AND fr.id in (SELECT fere.id
                                          FROM fee_reduction fere
                                                   JOIN finance_client fc on fere.finance_client_id = fc.client_id
                                          WHERE fc.client_id = $1)
                             AND i.id = $2)
UPDATE invoice i
SET fee_reduction_id = fi.fee_reduction_id
FROM filtered_invoices fi
WHERE i.id = fi.invoice_id
returning fi.fee_reduction_id, fi.type, fi.finance_client_id;

-- name: AddManualInvoice :one
INSERT INTO invoice (id, person_id, finance_client_id, feetype, reference, startdate, enddate, amount, confirmeddate,
                     batchnumber, raiseddate, source, scheduledfn14date, cacheddebtamount, createddate, createdby_id,
                     fee_reduction_id)
VALUES (nextval('invoice_id_seq'),
        $1,
        (select id from finance_client where client_id = $1),
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15)
returning *;

-- name: UpsertCounterForInvoiceRefYear :one
INSERT INTO counter (id, key, counter)
VALUES (nextval('counter_id_seq'), $1, 1)
ON CONFLICT (key) DO UPDATE
    SET counter = counter.counter + 1
RETURNING *;