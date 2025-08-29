-- +goose Up
CREATE TABLE pending_collection
(
    id                INT       NOT NULL PRIMARY KEY,
    finance_client_id INT       NOT NULL REFERENCES finance_client (id),
    collection_date   DATE      NOT NULL,
    amount            INT       NOT NULL,
    ledger_id         INT,
    created_at        TIMESTAMP NOT NULL,
    created_by        INT       NOT NULL
);

CREATE SEQUENCE pending_collection_id_seq;

-- +goose Down
DROP SEQUENCE pending_collection_id_seq;
DROP TABLE pending_collection;
