-- +goose Up
ALTER TABLE finance_client ADD COLUMN caserecnumber varchar(255);

-- +goose Down
ALTER TABLE finance_client DROP COLUMN caserecnumber varchar(255);