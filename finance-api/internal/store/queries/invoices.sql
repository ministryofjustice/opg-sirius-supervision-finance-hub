SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id,
       i.raiseddate,
       i.reference,
       i.amount,
       COALESCE(SUM(la.amount), 0)::INT    received,
       COALESCE(MAX(fr.type), '')::VARCHAR fee_reduction_type
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate
ORDER BY i.raiseddate DESC;

-- name: GetLedgerAllocations :many
WITH allocations AS (SELECT la.invoice_id,
                            la.amount,
                            COALESCE(l.bankdate, la.datetime) AS raised_date,
                            l.type,
                            la.status,
                            la.datetime AS created_at
                     FROM ledger_allocation la
                              JOIN ledger l ON la.ledger_id = l.id
                     WHERE la.invoice_id = ANY ($1::INT[])
                     UNION
                     SELECT ia.invoice_id, ia.amount, ia.raised_date, ia.adjustment_type, ia.status, ia.created_at
                     FROM invoice_adjustment ia
                     WHERE ia.status = 'PENDING'
                       AND ia.invoice_id = ANY ($1::INT[]))
SELECT *
FROM allocations
ORDER BY raised_date DESC, created_at DESC, status DESC;

-- name: GetSupervisionLevels :many
SELECT invoice_id, supervisionlevel, fromdate, todate, amount
FROM invoice_fee_range
WHERE invoice_id = ANY ($1::INT[])
ORDER BY todate DESC;

-- name: GetInvoiceBalanceDetails :one
SELECT i.amount                                                    initial,
       i.amount - COALESCE(SUM(la.amount), 0)                      outstanding,
       i.feetype,
       COALESCE(BOOL_OR(l.type = 'CREDIT WRITE OFF'), FALSE)::BOOL written_off
FROM invoice i
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id
         LEFT JOIN ledger l ON l.id = la.ledger_id
    AND la.status NOT IN ('PENDING', 'UNALLOCATED')
WHERE i.id = $1
GROUP BY i.amount, i.feetype;

-- name: GetInvoiceBalancesForFeeReductionRange :many
SELECT i.id,
       i.amount,
       ifr.amount AS                          general_supervision_fee,
       i.amount - COALESCE(SUM(la.amount), 0) outstanding,
       i.feetype
FROM invoice i
         JOIN fee_reduction fr ON i.finance_client_id = fr.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id
         LEFT JOIN ledger l ON l.id = la.ledger_id
         LEFT JOIN invoice_fee_range ifr ON i.id = ifr.invoice_id AND i.supervisionlevel = 'GENERAL'
WHERE i.raiseddate >= (fr.datereceived - INTERVAL '6 months')
  AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
  AND fr.id = $1
GROUP BY i.id, ifr.amount;

-- name: AddInvoice :one
INSERT INTO invoice (id, person_id, finance_client_id, feetype, reference, startdate, enddate, amount, confirmeddate,
                     raiseddate, source, created_at, created_by)
VALUES (NEXTVAL('invoice_id_seq'),
        $1,
        (SELECT id FROM finance_client WHERE client_id = $1),
        $2,
        $3,
        $4,
        $5,
        $6,
        NOW(),
        $7,
        $8,
        NOW(),
        $9)
RETURNING *;

-- name: GetInvoiceCounter :one
INSERT INTO counter (id, key, counter)
VALUES (NEXTVAL('counter_id_seq'), $1, 1)
ON CONFLICT (key) DO UPDATE
    SET counter = counter.counter + 1
RETURNING counter::VARCHAR;
