-- +goose Up
ALTER TABLE invoice ADD COLUMN fee_reduction_id integer;

-- +goose Down
ALTER TABLE invoice ADD COLUMN fee_reduction_id integer;
