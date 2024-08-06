SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id, i.reference, i.amount, i.raiseddate, COALESCE(SUM(la.amount), 0)::int received, fr.type fee_reduction_type
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate, fr.type
ORDER BY i.raiseddate DESC;

-- name: GetLedgerAllocations :many
SELECT la.invoice_id, la.id, la.amount, la.datetime, l.bankdate, l.type, la.status
FROM ledger_allocation la
         INNER JOIN ledger l ON la.ledger_id = l.id
WHERE la.invoice_id = ANY($1::int[])
ORDER BY la.id DESC;

-- name: GetSupervisionLevels :many
SELECT invoice_id, supervisionlevel, fromdate, todate, amount
FROM invoice_fee_range
WHERE invoice_id = ANY($1::int[])
ORDER BY todate DESC;

-- name: GetInvoiceBalanceDetails :one
SELECT i.amount initial, i.amount - COALESCE(SUM(la.amount), 0) outstanding, i.feetype,
       COALESCE(bool_or(l.type = 'CREDIT WRITE OFF'), false)::bool written_off
FROM invoice i
         LEFT JOIN ledger_allocation la on i.id = la.invoice_id
         LEFT JOIN ledger l ON l.id = la.ledger_id
    AND la.status NOT IN ('PENDING', 'UNALLOCATED')
WHERE i.id = $1
group by i.amount, i.feetype;

-- name: GetInvoiceBalancesForFeeReductionRange :many
SELECT i.id, i.amount, i.amount - COALESCE(SUM(la.amount), 0) outstanding, i.feetype
FROM invoice i
        JOIN fee_reduction fr ON i.finance_client_id = fr.finance_client_id
        LEFT JOIN ledger_allocation la on i.id = la.invoice_id
        LEFT JOIN ledger l ON l.id = la.ledger_id
WHERE i.raiseddate >= (fr.datereceived - interval '6 months')
 AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
 AND fr.id = $1
GROUP BY i.id;

-- name: AddInvoice :one
INSERT INTO invoice (id, person_id, finance_client_id, feetype, reference, startdate, enddate, amount, confirmeddate,
                     raiseddate, source, createddate, createdby_id)
VALUES (nextval('invoice_id_seq'),
        $1,
        (select id from finance_client where client_id = $1),
        $2,
        $3,
        $4,
        $5,
        $6,
        now(),
        $7,
        $8,
        now(),
        $9)
returning *;

-- name: GetInvoiceCounter :one
INSERT INTO counter (id, key, counter)
VALUES (nextval('counter_id_seq'), $1, 1)
ON CONFLICT (key) DO UPDATE
    SET counter = counter.counter + 1
RETURNING counter::VARCHAR;
