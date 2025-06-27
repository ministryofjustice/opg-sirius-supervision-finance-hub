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
UPDATE finance_client SET court_ref = $1 WHERE client_id = $2;

-- name: UpdatePaymentMethod :exec
UPDATE finance_client SET payment_method = $1 WHERE client_id = $2;

-- name: CheckClientExistsByCourtRef :one
SELECT EXISTS (SELECT 1 FROM finance_client WHERE court_ref = $1);

-- name: GetClientByCourtRef :one
SELECT id AS finance_client_id, client_id FROM finance_client WHERE court_ref = $1;
