-- +goose Up
ALTER TABLE refund RENAME client_id TO finance_client_id;
ALTER TABLE refund RENAME status TO decision;
ALTER TABLE refund RENAME updated_at TO decision_at;
ALTER TABLE refund RENAME updated_by TO decision_by;
ALTER TABLE refund ADD COLUMN processed_at TIMESTAMP;
ALTER TABLE refund ADD COLUMN cancelled_at TIMESTAMP;

ALTER TABLE refund ADD COLUMN fulfilled_at TIMESTAMP;
UPDATE refund SET fulfilled_at = fulfilled_date;
ALTER TABLE refund DROP COLUMN fulfilled_date;

-- +goose Down
ALTER TABLE refund RENAME decision TO status;
ALTER TABLE refund RENAME decision_at TO updated_at;
ALTER TABLE refund RENAME decision_by TO updated_by;
ALTER TABLE refund DROP COLUMN processed_at;
ALTER TABLE refund DROP COLUMN cancelled_at;

ALTER TABLE refund ADD COLUMN fulfilled_date DATE;
UPDATE refund SET fulfilled_date = fulfilled_at;
ALTER TABLE refund DROP COLUMN fulfilled_at;
