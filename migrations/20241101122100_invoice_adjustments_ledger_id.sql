-- +goose Up
ALTER TABLE invoice_adjustment
    ADD COLUMN ledger_id INTEGER
        CONSTRAINT ledger_id REFERENCES ledger (id);

-- +goose Down
-- +goose StatementBegin
ALTER TABLE invoice_adjustment
    DROP COLUMN ledger_id;
-- +goose StatementEnd
