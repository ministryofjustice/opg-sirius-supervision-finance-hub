SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoices :many
SELECT i.id,
       i.raiseddate,
       i.reference,
       i.amount,
       COALESCE(transactions.received, 0)::INT                AS received,
       COALESCE(transactions.fee_reduction_type, '')::VARCHAR AS fee_reduction_type
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN LATERAL (
    SELECT SUM(la.amount) AS received, MAX(fr.type) AS fee_reduction_type
    FROM ledger_allocation la
             JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
             LEFT JOIN fee_reduction fr ON l.fee_reduction_id = fr.id
    WHERE la.status NOT IN ('PENDING', 'UN ALLOCATED')
      AND la.invoice_id = i.id
    ) transactions ON TRUE
WHERE fc.client_id = $1
ORDER BY i.raiseddate DESC;

-- name: GetUnpaidInvoicesByCourtRef :many
SELECT i.id, (i.amount - COALESCE(transactions.received, 0)::INT) AS outstanding
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN LATERAL (
    SELECT SUM(la.amount) AS received
    FROM ledger_allocation la
             JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
    WHERE la.status NOT IN ('PENDING', 'UN ALLOCATED')
      AND la.invoice_id = i.id
    ) transactions ON TRUE
WHERE fc.court_ref = $1
GROUP BY i.id, i.amount, transactions.received, i.raiseddate
HAVING (i.amount - COALESCE(SUM(transactions.received), 0)::INT) > 0
ORDER BY i.raiseddate;

-- name: GetInvoicesForReversalByCourtRef :many
SELECT i.id, i.amount, (COALESCE(transactions.received, 0)::INT) AS received
FROM invoice i
         JOIN finance_client fc ON fc.id = i.finance_client_id
         LEFT JOIN LATERAL (
    SELECT SUM(la.amount) AS received
    FROM ledger_allocation la
             JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
    WHERE la.status NOT IN ('PENDING', 'UN ALLOCATED')
      AND la.invoice_id = i.id
    ) transactions ON TRUE
WHERE fc.court_ref = $1
GROUP BY i.id, i.amount, transactions.received, i.raiseddate
HAVING COALESCE(SUM(transactions.received), 0)::INT > 0
ORDER BY i.raiseddate DESC;

-- name: GetLedgerAllocations :many
WITH allocations AS (SELECT la.invoice_id,
                            la.amount,
                            COALESCE(l.bankdate, la.datetime) AS raised_date,
                            l.type,
                            la.status,
                            la.datetime                       AS created_at,
                            la.id                             AS ledger_allocation_id
                     FROM ledger_allocation la
                              JOIN ledger l ON la.ledger_id = l.id
                     WHERE la.invoice_id = ANY ($1::INT[])
                       AND la.status NOT IN ('PENDING', 'UN ALLOCATED')
                       AND l.status = 'CONFIRMED'
                     UNION
                     SELECT ia.invoice_id,
                            ia.amount,
                            ia.raised_date,
                            ia.adjustment_type,
                            ia.status,
                            ia.created_at,
                            ia.id
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
SELECT i.amount::INT                                        initial,
       (i.amount - COALESCE(transactions.received, 0))::INT outstanding,
       i.feetype,
       COALESCE(write_offs.amount, 0)::INT                  write_off_amount
FROM invoice i
         LEFT JOIN LATERAL (
    SELECT SUM(la.amount) AS amount
    FROM ledger_allocation la
             JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED' AND l.type = 'CREDIT WRITE OFF'
    WHERE la.status = 'ALLOCATED'
      AND la.invoice_id = i.id
    ) write_offs ON TRUE
         LEFT JOIN LATERAL (
    SELECT SUM(la.amount) AS received
    FROM ledger_allocation la
             JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
    WHERE la.status NOT IN ('PENDING', 'UN ALLOCATED')
      AND la.invoice_id = i.id
    ) transactions ON TRUE
WHERE i.id = $1;

-- name: GetInvoiceBalancesForFeeReductionRange :many
SELECT i.id,
       i.amount,
       COALESCE(general_fee.amount, 0)::INT               general_supervision_fee,
       i.amount - COALESCE(transactions.received, 0)::INT outstanding,
       i.feetype
FROM invoice i
         JOIN fee_reduction fr ON i.finance_client_id = fr.finance_client_id
         LEFT JOIN LATERAL (
    SELECT SUM(la.amount) AS received
    FROM ledger_allocation la
             JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
    WHERE la.status NOT IN ('PENDING', 'UN ALLOCATED')
      AND la.invoice_id = i.id
    ) transactions ON TRUE
         LEFT JOIN LATERAL (
    SELECT SUM(ifr.amount) AS amount
    FROM invoice_fee_range ifr
    WHERE ifr.invoice_id = i.id
      AND ifr.supervisionlevel = 'GENERAL'
    ) general_fee ON TRUE
WHERE i.raiseddate >= (fr.datereceived - INTERVAL '6 months')
  AND i.raiseddate BETWEEN fr.startdate AND fr.enddate
  AND fr.id = $1;

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
