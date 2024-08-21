-- +goose Up
ALTER TABLE invoice ALTER COLUMN createddate TYPE timestamp(0);
ALTER TABLE invoice ALTER COLUMN confirmeddate TYPE timestamp(0);
ALTER TABLE fee_reduction ALTER COLUMN created_at TYPE timestamp(0);
ALTER TABLE fee_reduction ALTER COLUMN cancelled_at TYPE timestamp(0);

-- +goose Down
ALTER TABLE invoice ALTER COLUMN createddate TYPE date;
ALTER TABLE invoice ALTER COLUMN confirmeddate TYPE date;
ALTER TABLE fee_reduction ALTER COLUMN created_at TYPE date;
ALTER TABLE fee_reduction ALTER COLUMN cancelled_at TYPE date;
