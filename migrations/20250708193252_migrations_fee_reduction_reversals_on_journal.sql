-- +goose Up
INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description, line_description, is_receipt)
VALUES
       ('FRR', 'AD', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'AD Fee reduction reversal', false);


-- +goose Down
