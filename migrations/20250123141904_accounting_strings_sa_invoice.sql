-- +goose Up
INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description)
VALUES
    ('SA', '', '', 4481102093, 'SA - General invoice (Legacy)');

-- +goose Down

DELETE FROM transaction_type WHERE fee_type = 'SA';
