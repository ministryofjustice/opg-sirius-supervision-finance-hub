-- +goose Up
ALTER TABLE ledger ADD COLUMN createddate date;
ALTER TABLE ledger ADD COLUMN createdby_id int;

-- +goose Down
ALTER TABLE ledger DROP COLUMN createddate;
ALTER TABLE ledger DROP COLUMN createdby_id;
