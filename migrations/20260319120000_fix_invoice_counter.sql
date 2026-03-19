-- +goose Up
-- Due to a race condition, the invoice counter may have been incremented only once when two invoices
-- were created simultaneously, leaving the counter behind the highest reference already in use. This
-- migration resyncs all existing counter rows to the actual max reference number in the invoice table,
-- so that the next generated reference (max + 1) is guaranteed to be unique.
-- The counter key is derived from the calendar year of the invoice start date, matching the application logic in generateInvoiceReference.
UPDATE counter
SET counter = subquery.max_counter
FROM (
    SELECT
        EXTRACT(YEAR FROM startdate)::INTEGER::TEXT || 'InvoiceNumber' AS key,
        MAX(CAST(SUBSTRING(SPLIT_PART(reference, '/', 1), 3) AS INTEGER)) AS max_counter
    FROM invoice
    GROUP BY EXTRACT(YEAR FROM startdate)
) subquery
WHERE counter.key = subquery.key
  AND counter.counter < subquery.max_counter;

-- +goose Down

