-- +goose Up
ALTER TABLE ledger_allocation ADD COLUMN transaction_type varchar(255);

-- +goose Down
ALTER TABLE ledger_allocation DROP COLUMN transaction_type;