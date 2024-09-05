-- +goose Up
ALTER TABLE invoice ADD COLUMN created_at timestamp(0);
ALTER TABLE invoice ADD COLUMN created_by integer;

-- +goose Down
ALTER TABLE invoice DROP COLUMN created_at;
ALTER TABLE invoice DROP COLUMN created_by;
