-- +goose Up
INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description, line_description, is_receipt)
VALUES
       ('FRR', 'AD', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'AD Fee reduction reversal', false),
       ('FRR', 'GENERAL', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'General Fee reduction reversal', false),
       ('FRR', 'MINIMAL', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'Minimal Fee reduction reversal', false),
       ('FRR', 'GA', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'GA Fee reduction reversal', false),
       ('FRR', 'GT', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'GT Fee reduction reversal', false),
       ('FRR', 'GS', 'FEE REDUCTION REVERSAL', 4481102107, 'Fee Reduction Reversal', 'GS Fee reduction reversal', false);


-- +goose Down
DELETE FROM transaction_type WHERE fee_type = 'FRR';