-- +goose Up
ALTER TABLE supervision_finance.invoice ADD COLUMN createddate date;
ALTER TABLE supervision_finance.invoice ADD COLUMN createdby_id integer;

-- +goose Down
ALTER TABLE supervision_finance.invoice DROP COLUMN createddate;
ALTER TABLE supervision_finance.invoice DROP COLUMN createdby_id;
