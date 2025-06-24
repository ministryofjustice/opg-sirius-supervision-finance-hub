-- +goose Up
ALTER TABLE refund ADD COLUMN cancelled_by INT;

-- +goose Down
ALTER TABLE refund DROP COLUMN cancelled_by;

