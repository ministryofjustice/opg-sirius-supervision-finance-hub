-- +goose Up
DROP SEQUENCE IF EXISTS report_id_seq CASCADE;
DROP TABLE IF EXISTS report;

-- +goose Down
-- no down migration for dropping a table