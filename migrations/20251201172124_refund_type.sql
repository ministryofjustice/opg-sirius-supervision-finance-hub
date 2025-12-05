-- +goose Up
UPDATE transaction_type SET ledger_type = 'REFUND', line_description = 'Refund', account_code = '1841102050' WHERE fee_type = 'BCR';

-- +goose Down
UPDATE transaction_type SET ledger_type = '', line_description = 'BACS Refund', account_code = '1816102006' WHERE fee_type = 'BCR';
