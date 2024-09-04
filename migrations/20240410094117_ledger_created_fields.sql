-- +goose Up
ALTER TABLE ledger ADD COLUMN created_at timestamp(0);
ALTER TABLE ledger ADD COLUMN created_by int;

-- +goose Down
ALTER TABLE ledger DROP COLUMN created_at;
ALTER TABLE ledger DROP COLUMN created_by;
