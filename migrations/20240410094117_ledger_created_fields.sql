-- +goose Up
ALTER TABLE supervision_finance.ledger ADD COLUMN createddate date;
ALTER TABLE supervision_finance.ledger ADD COLUMN createdby_id int;

-- +goose Down
ALTER TABLE supervision_finance.ledger DROP COLUMN createddate;
ALTER TABLE supervision_finance.ledger DROP COLUMN createdby_id;
