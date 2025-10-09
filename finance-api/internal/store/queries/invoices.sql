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
                            la.datetime AS received_date,
                            l.type,
                            la.status,
                            la.datetime AS created_at,
                            la.id       AS ledger_allocation_id
                     FROM ledger_allocation la
                              JOIN ledger l ON la.ledger_id = l.id
                     WHERE la.invoice_id = ANY ($1::INT[])
                       AND la.status NOT IN ('PENDING', 'UN ALLOCATED')
                       AND l.status = 'CONFIRMED'
                     UNION
                     SELECT ia.invoice_id,
                            ia.amount,
                            ia.raised_date AS received_date,
                            ia.adjustment_type,
                            ia.status,
                            ia.created_at,
                            ia.id
                     FROM invoice_adjustment ia
                     WHERE ia.status = 'PENDING'
                       AND ia.invoice_id = ANY ($1::INT[]))
SELECT *
FROM allocations
ORDER BY received_date DESC, created_at DESC, status DESC, ledger_allocation_id ASC;

-- name: GetSupervisionLevels :many
SELECT invoice_id, supervisionlevel, fromdate, todate, amount
FROM invoice_fee_range
WHERE invoice_id = ANY ($1::INT[])
ORDER BY todate DESC;

-- name: GetInvoiceBalanceDetails :one
WITH ledger_sums AS (SELECT la.invoice_id,
                            SUM(CASE
                                    WHEN l.status = 'CONFIRMED' AND l.type = 'CREDIT WRITE OFF' AND
                                         la.status = 'ALLOCATED' THEN la.amount
                                    ELSE 0 END) AS write_off_amount,
                            SUM(CASE
                                    WHEN l.status = 'CONFIRMED' AND l.type = 'WRITE OFF REVERSAL' AND
                                         la.status = 'ALLOCATED' THEN la.amount
                                    ELSE 0 END) AS write_off_reversal_amount,
                            SUM(CASE
                                    WHEN l.status = 'CONFIRMED' AND la.status NOT IN ('PENDING', 'UN ALLOCATED')
                                        THEN la.amount
                                    ELSE 0 END) AS received
                     FROM ledger_allocation la
                              JOIN ledger l ON la.ledger_id = l.id
                     GROUP BY la.invoice_id)
SELECT i.amount::INT                                                                          AS initial,
       i.amount - COALESCE(ls.received, 0)::INT                                               AS outstanding,
       i.feetype,
       COALESCE(ls.write_off_amount, 0)::INT + COALESCE(ls.write_off_reversal_amount, 0)::INT AS write_off_amount
FROM invoice i
         LEFT JOIN ledger_sums ls ON ls.invoice_id = i.id
WHERE i.id = @invoice_id;

-- name: GetInvoiceFeeReductionReversalDetails :one
SELECT (SELECT COALESCE(SUM(amount), 0)
        FROM invoice_adjustment ia
        WHERE ia.invoice_id = @invoice_id
          AND ia.adjustment_type = 'FEE REDUCTION REVERSAL'
          AND ia.status = 'APPROVED')         AS reversal_total,
       (SELECT COALESCE(SUM(la.amount), 0)
        FROM ledger l
                 JOIN ledger_allocation la ON l.id = la.ledger_id
        WHERE la.invoice_id = @invoice_id
          AND la.status = 'ALLOCATED'
          AND l.fee_reduction_id IS NOT NULL) AS fee_reduction_total;

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
