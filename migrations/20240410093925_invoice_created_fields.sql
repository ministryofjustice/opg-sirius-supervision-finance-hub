-- +goose Up
ALTER TABLE invoice ADD COLUMN createddate date;
ALTER TABLE invoice ADD COLUMN createdby_id integer;

-- +goose Down
ALTER TABLE invoice DROP COLUMN createddate;
ALTER TABLE invoice DROP COLUMN createdby_id;
