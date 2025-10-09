-- +goose Up
ALTER TABLE transaction_type ADD COLUMN line_description VARCHAR, ADD COLUMN is_receipt BOOLEAN;

CREATE TEMP TABLE transaction_type_update(
    transaction_type_id INT PRIMARY KEY,
    line_description_update VARCHAR,
    is_receipt_update BOOLEAN
);

INSERT INTO transaction_type_update VALUES
(1, 'AD invoice', false),
(2, 'S2 invoice', false),
(3, 'S3 invoice', false),
(4, 'B2 invoice', false),
(5, 'B3 invoice', false),
(6, 'Gen SF invoice', false),
(7, 'Min SF invoice', false),
(8, 'Gen SE invoice', false),
(9, 'Min SE invoice', false),
(10, 'Gen SO invoice', false),
(11, 'Min SO invoice', false),
(12, 'AD Rem/Exem', false),
(13, 'AD Rem/Exem', false),
(14, 'AD Rem/Exem', false),
(15, 'AD Manual credit', false),
(16, 'AD Manual debit', false),
(17, 'AD Write-off', false),
(18, 'AD Write-off reversal', false),
(19, 'Gen Rem/Exem', false),
(20, 'Gen Rem/Exem', false),
(21, 'Gen Rem/Exem', false),
(22, 'Gen Manual credit', false),
(23, 'Gen Manual debit', false),
(24, 'Gen Write-off', false),
(25, 'Gen Write-off reversal', false),
(26, 'Min Rem/Exem', false),
(27, 'Min Rem/Exem', false),
(28, 'Min Rem/Exem', false),
(29, 'Min Manual credit', false),
(30, 'Min Manual debit', false),
(31, 'Min Write-off', false),
(32, 'Min Write-off reversal', false),
(33, 'GA invoice', false),
(34, 'GS invoice', false),
(35, 'GT invoice', false),
(36, 'Unapply (money from invoice)', false),
(37, 'Overpayment', false),
(38, 'Cheque Refund', false),
(39, 'BACS Refund', false),
(40, 'Direct debit', true),
(41, 'Online card', true),
(42, 'MOTO card', true),
(43, 'Supervision BACS', true),
(44, 'OPG BACS', true),
(45, 'Cheque payment', true),
(46, 'Reapply/Reallocate (money to invoice)', true),
(47, 'BACS Payment', true),
(48, 'Online card', true);

UPDATE transaction_type
SET line_description = line_description_update, is_receipt = is_receipt_update
FROM transaction_type_update
WHERE transaction_type_id = id;

UPDATE transaction_type SET supervision_level = fee_type WHERE supervision_level = '' AND fee_type IN ('GA', 'GS', 'GT');

-- +goose Down

ALTER TABLE transaction_type DROP COLUMN line_description, DROP COLUMN is_receipt;
