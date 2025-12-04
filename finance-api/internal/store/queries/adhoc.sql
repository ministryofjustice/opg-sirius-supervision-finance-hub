-- name: UpdateRefundLedgerAmounts :execrows
UPDATE ledger
SET amount = -amount
WHERE type = 'REFUND'
  AND amount > 0;
