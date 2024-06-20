-- name: GetInvoiceFeeRangeAmount :one
SELECT SUM(amount) / 2
FROM supervision_finance.invoice_fee_range
WHERE invoice_id = $1
  AND supervisionlevel = $2;