-- +goose Up
ALTER TABLE ledger ADD COLUMN created_at date;
ALTER TABLE ledger ADD COLUMN created_by int;

-- +goose Down
ALTER TABLE ledger DROP COLUMN created_at;
ALTER TABLE ledger DROP COLUMN created_by;
