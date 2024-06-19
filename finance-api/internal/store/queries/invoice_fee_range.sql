SET SEARCH_PATH TO supervision_finance;

-- name: GetInvoiceFeeRangeAmount :one
select sum(amount) / 2
from invoice_fee_range
where invoice_id = $1
  and supervisionlevel = $2;

-- name: AddInvoiceRange :exec
INSERT INTO invoice_fee_range (id, invoice_id, supervisionlevel, fromdate, todate, amount)
VALUES (nextval('invoice_fee_range_id_seq'),
        $1,
        $2,
        $3,
        $4,
        $5);