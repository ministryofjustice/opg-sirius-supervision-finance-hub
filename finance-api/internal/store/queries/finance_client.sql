-- name: GetAccountInformation :one
SELECT cacheddebtamount, cachedcreditamount, payment_method FROM finance_client WHERE client_id = $1;

-- name: GetInvoices :many
SELECT id, reference, amount, raiseddate, cacheddebtamount FROM invoice WHERE finance_client_id = $1;