-- +goose Up
ALTER TABLE finance_client ADD COLUMN cacheddebtamount integer;
ALTER TABLE finance_client ADD COLUMN cachedcreditamount integer;

-- +goose Down
ALTER TABLE finance_client DROP COLUMN cacheddebtamount;
ALTER TABLE finance_client DROP COLUMN cachedcreditamount;
