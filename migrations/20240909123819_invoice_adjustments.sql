-- +goose Up
CREATE TABLE invoice_adjustments (
    id INTEGER NOT NULL PRIMARY KEY,
    client_id INTEGER NOT NULL REFERENCES finance_client,
    invoice_id INTEGER NOT NULL REFERENCES invoice,
    raised_date DATE NOT NULL,
    adjustment_type VARCHAR(255) NOT NULL,
    amount INTEGER NOT NULL,
    notes TEXT NOT NULL,
    status VARCHAR(50) NOT NULL,
    created_at TIMESTAMP(0) NOT NULL,
    created_by INTEGER NOT NULL,
    updated_at TIMESTAMP(0),
    updated_by INTEGER
);

create sequence invoice_adjustments_id_seq;

-- +goose Down
DROP TABLE invoice_adjustments;
