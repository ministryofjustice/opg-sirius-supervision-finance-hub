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
         LEFT JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.client_id = $1
GROUP BY i.id, i.raiseddate
ORDER BY i.raiseddate DESC;

-- name: GetInvoicesForCourtRef :many
SELECT i.id,
       (i.amount - COALESCE(SUM(la.amount), 0)::INT) outstanding
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
         LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
WHERE fc.court_ref = $1
GROUP BY i.id, i.raiseddate
HAVING (i.amount - COALESCE(SUM(la.amount), 0)::INT) > 0
ORDER BY i.raiseddate ASC;

-- name: GetLedgerAllocations :many
WITH allocations AS (SELECT la.invoice_id,
                            la.amount,
                            COALESCE(l.bankdate, la.datetime) AS raised_date,
                            l.type,
                            la.status,
                            la.datetime AS created_at,
                            la.id AS ledger_allocation_id
                     FROM ledger_allocation la
                              JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
                     WHERE la.invoice_id = ANY ($1::INT[])
                     UNION
                     SELECT ia.invoice_id, ia.amount, ia.raised_date, ia.adjustment_type, ia.status, ia.created_at, ia.id
                     FROM invoice_adjustment ia
                     WHERE ia.status = 'PENDING'
                       AND ia.invoice_id = ANY ($1::INT[]))
SELECT *
FROM allocations
ORDER BY raised_date DESC, created_at DESC, status DESC, ledger_allocation_id ASC;

-- name: GetSupervisionLevels :many
SELECT invoice_id, supervisionlevel, fromdate, todate, amount
FROM invoice_fee_range
WHERE invoice_id = ANY ($1::INT[])
ORDER BY todate DESC;

-- name: GetInvoiceBalanceDetails :one
SELECT i.amount                                                    initial,
       i.amount - COALESCE(SUM(la.amount), 0)                      outstanding,
       i.feetype,
       COALESCE((SELECT SUM(ledger_allocation.amount) FROM ledger_allocation LEFT JOIN ledger ON ledger_allocation.ledger_id = ledger.id LEFT JOIN invoice ON ledger_allocation.invoice_id = invoice.id WHERE ledger.type = 'CREDIT WRITE OFF' AND invoice.id = i.id), 0)::INT write_off_amount
FROM invoice i
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
WHERE i.id = $1
GROUP BY i.amount, i.feetype, i.id;

-- name: GetInvoiceBalancesForFeeReductionRange :many
SELECT i.id,
       i.amount,
       COALESCE(general_fee.amount, 0)                          general_supervision_fee,
       i.amount - COALESCE(SUM(la.amount), 0) outstanding,
       i.feetype
FROM invoice i
         JOIN fee_reduction fr ON i.finance_client_id = fr.finance_client_id
         LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
         LEFT JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
         LEFT JOIN LATERAL (
             SELECT SUM(ifr.amount) AS amount
             FROM invoice_fee_range ifr
             WHERE ifr.invoice_id = i.id
             AND ifr.supervisionlevel = 'GENERAL'
         ) general_fee ON TRUE
WHERE i.raiseddate >= (fr.datereceived - INTERVAL '6 months')
  AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
  AND fr.id = $1
GROUP BY i.id, general_fee.amount;

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
