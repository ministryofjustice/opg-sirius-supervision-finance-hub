-- +goose Up
CREATE TABLE payment_method
(
    id                INTEGER NOT NULL PRIMARY KEY,
    finance_client_id INT NOT NULL REFERENCES finance_client (id),
    type              VARCHAR NOT NULL,
    created_at        TIMESTAMP NOT NULL,
    created_by        INTEGER NOT NULL
);

CREATE SEQUENCE payment_method_id_seq;

-- +goose Down
DROP SEQUENCE payment_method_id_seq;
DROP TABLE payment_method;