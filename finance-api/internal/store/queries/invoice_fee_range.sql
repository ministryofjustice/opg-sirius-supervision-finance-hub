SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoiceFeeRangeAmount :one
select sum(amount) / 2
from invoice_fee_range
where invoice_id = $1 and supervisionlevel = $2;