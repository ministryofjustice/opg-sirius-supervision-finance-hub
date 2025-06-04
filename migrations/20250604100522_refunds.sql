-- +goose Up
CREATE TABLE refund
(
    id             INT NOT NULL PRIMARY KEY,
    client_id      INT NOT NULL REFERENCES finance_client(id),
    raised_date    DATE NOT NULL,
    fulfilled_date DATE DEFAULT NULL,
    amount         INT NOT NULL,
    status         VARCHAR NOT NULL,
    notes          VARCHAR NOT NULL,
    created_by     INT NOT NULL,
    created_at     TIMESTAMP NOT NULL,
    updated_by     INT DEFAULT NULL,
    updated_at     TIMESTAMP DEFAULT NULL
);

CREATE SEQUENCE refund_id_seq;

CREATE TABLE bank_details
(
    id INT NOT NULL PRIMARY KEY,
    refund_id INT NOT NULL REFERENCES refund(id),
    name VARCHAR NOT NULL,
    account VARCHAR NOT NULL,
    sort_code VARCHAR NOT NULL
);

CREATE SEQUENCE bank_details_id_seq;

-- +goose Down
DROP SEQUENCE bank_details_id_seq;
DROP TABLE bank_details;

DROP SEQUENCE refund_id_seq;
DROP TABLE refund;
