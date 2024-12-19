-- +goose Up
ALTER TABLE transaction_type
ALTER COLUMN supervision_level DROP NOT NULL;

UPDATE transaction_type
SET account_code = CASE
                       WHEN fee_type = 'MCR' AND supervision_level = 'AD' THEN 4481102114
                       WHEN fee_type = 'MCR' AND supervision_level = 'GENERAL' THEN 4481102115
                       WHEN fee_type = 'MCR' AND supervision_level = 'MINIMAL' THEN 4481102120
                       WHEN fee_type = 'MDR' AND supervision_level = 'AD' THEN 4481102114
                       WHEN fee_type = 'MDR' AND supervision_level = 'GENERAL' THEN 4481102115
                       WHEN fee_type = 'MDR' AND supervision_level = 'MINIMAL' THEN 4481102120
                       ELSE account_code
    END
WHERE fee_type IN ('MCR', 'MDR');

UPDATE transaction_type
SET supervision_level = NULL WHERE supervision_level = '';

INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description)
VALUES
('DD', NULL, 'DIRECT DEBIT PAYMENT', 1816100000, 'Direct Debit Payment'),
('OC', NULL, 'ONLINE CARD PAYMENT', 1816100000, 'Online Card Payment'),
('PC', NULL, 'MOTO CARD PAYMENT', 1816100000, 'MOTO (phone) Card Payment'),
('BC', NULL, 'SUPERVISION BACS PAYMENT', 1816100000, 'BACS Payment'),
('BC', NULL, 'OPG BACS PAYMENT', 1816100000, 'BACS Payment'),
('CQ', NULL, 'CHEQUE PAYMENT', 1816100000, 'Cheque Payment'),
('RA', NULL, 'CREDIT REAPPLY', 1816100000, 'Reapply/Reallocate (money to invoice)'),
('BC', NULL, 'BACS TRANSFER', 1816100000, 'BACS Payment'),
('PC', NULL, 'CARD PAYMENT', 1816100000, 'Card Payment');

-- +goose Down

DELETE FROM transaction_type WHERE fee_type IN ('DD', 'OC', 'PC', 'BC', 'CQ', '');

UPDATE transaction_type
SET account_code = CASE
                       WHEN fee_type = 'MCR' AND supervision_level = 'AD' THEN 4481102093
                       WHEN fee_type = 'MCR' AND supervision_level = 'GENERAL' THEN 4481102094
                       WHEN fee_type = 'MCR' AND supervision_level = 'MINIMAL' THEN 4481102099
                       WHEN fee_type = 'MDR' AND supervision_level = 'AD' THEN 4481102093
                       WHEN fee_type = 'MDR' AND supervision_level = 'GENERAL' THEN 4481102094
                       WHEN fee_type = 'MDR' AND supervision_level = 'MINIMAL' THEN 4481102099
                       ELSE account_code
                   END
WHERE fee_type IN ('MCR', 'MDR');

UPDATE transaction_type
SET supervision_level = '' WHERE supervision_level IS NULL;

ALTER TABLE transaction_type
    ALTER COLUMN supervision_level SET NOT NULL;