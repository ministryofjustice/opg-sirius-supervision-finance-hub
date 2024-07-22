-- +goose Up
ALTER TABLE fee_reduction ADD COLUMN created_at date;
ALTER TABLE fee_reduction ADD COLUMN created_by int;
ALTER TABLE fee_reduction ADD COLUMN cancelled_at date;
ALTER TABLE fee_reduction ADD COLUMN cancelled_by int;
ALTER TABLE fee_reduction ADD COLUMN cancellation_reason TEXT;

-- +goose Down
ALTER TABLE fee_reduction DROP COLUMN created_at;
ALTER TABLE fee_reduction DROP COLUMN created_by;
ALTER TABLE fee_reduction DROP COLUMN cancelled_at;
ALTER TABLE fee_reduction DROP COLUMN cancelled_by;
ALTER TABLE fee_reduction DROP COLUMN cancellation_reason;
