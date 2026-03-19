-- +goose Up
-- The invoice counter for the current year may be out of sync with the actual invoices in the database,
-- causing duplicate reference violations on insert. This migration resets the counter to the highest
-- reference number already in use for each year key. The application always increments the counter
-- before using it to generate a reference, so the next reference produced will be max + 1, guaranteed
-- to be unique.
UPDATE counter
SET counter = subquery.max_counter
FROM (
    SELECT
        EXTRACT(YEAR FROM startdate)::TEXT || 'InvoiceNumber' AS key,
        MAX(CAST(SUBSTRING(reference, 3, 6) AS INTEGER))      AS max_counter
    FROM invoice
    GROUP BY EXTRACT(YEAR FROM startdate)
) subquery
WHERE counter.key = subquery.key;

-- +goose Down

