-- name: GetCreditBalanceAndOldestOpenInvoice :one
WITH total_credit AS (SELECT fc.id                                 AS finance_client_id,
                             ABS(COALESCE(SUM(la.amount), 0))::INT AS credit
                      FROM finance_client fc
                               LEFT JOIN ledger l ON fc.id = l.finance_client_id
                               LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
                      WHERE fc.client_id = $1
                        AND la.status IN ('UNAPPLIED', 'REAPPLIED')
                      GROUP BY fc.id),
     oldest_unpaid AS (SELECT i.finance_client_id,
                              i.id                                   AS invoice_id,
                              i.amount - COALESCE(SUM(la.amount), 0) AS outstanding
                       FROM invoice i
                                LEFT JOIN ledger_allocation la ON i.id = la.invoice_id AND la.status NOT IN ('PENDING', 'UNALLOCATED')
                                LEFT JOIN ledger l ON la.ledger_id = l.id AND l.status = 'CONFIRMED'
                       WHERE i.finance_client_id = (SELECT fc.id
                                                    FROM finance_client fc
                                                    WHERE fc.client_id = $1)
                       GROUP BY i.id, i.raiseddate, i.amount
                       HAVING (i.amount - COALESCE(SUM(la.amount), 0)) > 0 -- Only unpaid invoices
                       ORDER BY i.raiseddate
                       LIMIT 1)
SELECT tc.credit,
       ou.invoice_id,
       ou.outstanding
FROM total_credit tc
         LEFT JOIN oldest_unpaid ou ON tc.finance_client_id = ou.finance_client_id;
