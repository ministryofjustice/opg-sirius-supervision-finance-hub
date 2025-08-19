-- +goose Up
ALTER TABLE ledger ADD COLUMN general_ledger_date DATE;
UPDATE ledger SET general_ledger_date = created_at::DATE WHERE created_at IS NOT NULL;

-- +goose Down
ALTER TABLE ledger DROP COLUMN general_ledger_date;
