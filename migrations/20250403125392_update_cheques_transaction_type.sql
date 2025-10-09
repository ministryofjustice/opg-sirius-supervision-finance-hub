-- +goose Up
UPDATE transaction_type SET ledger_type = 'SUPERVISION CHEQUE PAYMENT' WHERE ledger_type = 'CHEQUE PAYMENT';
-- +goose Down
UPDATE transaction_type SET ledger_type = 'CHEQUE PAYMENT' WHERE ledger_type = 'SUPERVISION CHEQUE PAYMENT';