-- +goose Up
CREATE TABLE pending_collection
(
    id                INTEGER   NOT NULL PRIMARY KEY,
    finance_client_id INTEGER REFERENCES finance_client (id),
    collection_date   DATE      NOT NULL,
    amount            INTEGER   NOT NULL,
    status            VARCHAR   NOT NULL,
    ledger_id         INTEGER REFERENCES ledger (id),
    created_at        TIMESTAMP NOT NULL,
    created_by        INTEGER   NOT NULL
);

CREATE INDEX idx_pending_collection_client_id ON pending_collection (finance_client_id);
CREATE INDEX idx_pending_collection_status ON pending_collection (status);
CREATE SEQUENCE pending_collection_id_seq;

-- +goose Down
DROP INDEX idx_pending_collection_client_id;
DROP INDEX idx_pending_collection_status;
DROP SEQUENCE pending_collection_id_seq;
DROP TABLE pending_collection;
