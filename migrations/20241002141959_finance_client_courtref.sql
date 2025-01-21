-- +goose Up
ALTER TABLE finance_client ADD COLUMN IF NOT EXISTS court_ref varchar(255);
CREATE INDEX IF NOT EXISTS idx_finance_client_court_ref ON finance_client(court_ref);

-- +goose Down
ALTER TABLE finance_client DROP COLUMN court_ref varchar(255);
DROP INDEX idx_finance_client_court_ref;