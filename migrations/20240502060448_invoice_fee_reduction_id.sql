-- +goose Up
ALTER TABLE supervision_finance.invoice ADD COLUMN fee_reduction_id integer;

-- +goose Down
ALTER TABLE supervision_finance.invoice ADD COLUMN fee_reduction_id integer;
