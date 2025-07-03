-- name: GetAccountInformation :one
WITH balances AS (SELECT fc.id,
                         COALESCE(SUM(
                                          CASE
                                              WHEN la.status = 'ALLOCATED'
                                                  THEN la.amount
                                              WHEN la.status IN ('UNAPPLIED', 'REAPPLIED') AND la.invoice_id IS NOT NULL
                                                  THEN la.amount
                                              ELSE 0
                                              END), 0)::INT      AS paid,
                         ABS(COALESCE(SUM(
                                              CASE
                                                  WHEN la.status IN ('UNAPPLIED', 'REAPPLIED')
                                                      THEN la.amount
                                                  ELSE 0
                                                  END), 0))::INT AS credit
                  FROM finance_client fc
                           LEFT JOIN ledger l ON fc.id = l.finance_client_id AND l.status = 'CONFIRMED'
                           LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
                  WHERE fc.client_id = $1
                  GROUP BY fc.id)
SELECT COALESCE(SUM(i.amount), 0)::INT - b.paid AS outstanding,
       b.credit,
       fc.payment_method
FROM finance_client fc
         JOIN balances b ON fc.id = b.id
         LEFT JOIN invoice i ON fc.id = i.finance_client_id
GROUP BY fc.payment_method, b.paid, b.credit;

-- name: UpdateClient :exec
UPDATE finance_client
SET court_ref = $1
WHERE client_id = $2;

-- name: UpdatePaymentMethod :exec
UPDATE finance_client
SET payment_method = $1
WHERE client_id = $2;

-- name: CheckClientExistsByCourtRef :one
SELECT EXISTS (SELECT 1 FROM finance_client WHERE court_ref = $1);

-- name: GetClientByCourtRef :one
SELECT id AS finance_client_id, client_id FROM finance_client WHERE court_ref = $1;


-- name: GetReversibleBalanceByCourtRef :one
WITH paid_on_invoices AS (SELECT fc.id, SUM(COALESCE(transactions.received, 0))::INT AS received
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
                            AND transactions.received > 0
                          GROUP BY fc.id),
     credit_balance AS (SELECT fc.id,
                               ABS(COALESCE(SUM(
                                                    CASE
                                                        WHEN la.status IN ('UNAPPLIED', 'REAPPLIED')
                                                            THEN la.amount
                                                        ELSE 0
                                                        END), 0))::INT AS credit
                        FROM finance_client fc
                                 LEFT JOIN ledger l ON fc.id = l.finance_client_id AND l.status = 'CONFIRMED'
                                 LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
                        WHERE fc.court_ref = $1
                        GROUP BY fc.id)
SELECT COALESCE(poi.received + cb.credit, 0) AS balance
FROM finance_client fc
         LEFT JOIN paid_on_invoices poi ON fc.id = poi.id
         LEFT JOIN credit_balance cb ON poi.id = cb.id
WHERE fc.court_ref = $1;

-- name: GetCreditBalanceByCourtRef :one
SELECT ABS(COALESCE(SUM(la.amount), 0))::INT AS credit
FROM finance_client fc
         LEFT JOIN ledger l ON fc.id = l.finance_client_id
         LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
WHERE fc.court_ref = $1
  AND la.status IN ('UNAPPLIED', 'REAPPLIED');
