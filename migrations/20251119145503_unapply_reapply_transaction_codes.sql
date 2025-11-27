-- +goose Up
UPDATE transaction_type SET account_code = 1816102003, line_description = 'Unapply' WHERE fee_type = 'UA';
UPDATE transaction_type SET line_description = 'Reapply' WHERE fee_type = 'RA';

-- +goose Down
UPDATE transaction_type SET account_code = 1816102004, line_description = 'Unapply (money from invoice)' WHERE fee_type = 'UA';
UPDATE transaction_type SET line_description = 'Reapply/Reallocate (money to invoice)' WHERE fee_type = 'RA';
