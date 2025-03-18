-- +goose Up
ALTER TABLE ledger ADD COLUMN pis_number INTEGER;

-- +goose Down

ALTER TABLE ledger DROP COLUMN pis_number;