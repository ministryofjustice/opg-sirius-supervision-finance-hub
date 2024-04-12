-- +goose Up
ALTER TABLE invoice ADD COLUMN createddate date;
ALTER TABLE invoice ADD COLUMN createdby_id integer;

-- +goose Down
ALTER TABLE invoice DROP COLUMN person_id;
ALTER TABLE invoice DROP COLUMN supervisionlevel;
ALTER TABLE invoice DROP COLUMN confirmeddate;
ALTER TABLE invoice DROP COLUMN batchnumber;
ALTER TABLE invoice DROP COLUMN createddate;
ALTER TABLE invoice DROP COLUMN createdby_id;
