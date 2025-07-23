-- +goose Up
DROP SEQUENCE report_id_seq CASCADE;
DROP TABLE report;

-- +goose Down
-- no down migration for dropping a table