-- name: GetAccountInformation :one
SELECT cacheddebtamount, cachedcreditamount, payment_method FROM finance_client WHERE client_id = $1;