-- +goose Up
ALTER TABLE supervision_finance.ledger_allocation ADD COLUMN transaction_type varchar(255);

-- +goose Down
ALTER TABLE supervision_finance.ledger_allocation DROP COLUMN transaction_type;