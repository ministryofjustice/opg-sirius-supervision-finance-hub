-- +goose Up
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

INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description)
VALUES
('DD', '', 'DIRECT DEBIT PAYMENT', 1816100000, 'Direct Debit Payment'),
('OC', '', 'ONLINE CARD PAYMENT', 1816100000, 'Online Card Payment'),
('PC', '', 'MOTO CARD PAYMENT', 1816100000, 'MOTO (phone) Card Payment'),
('BC', '', 'SUPERVISION BACS PAYMENT', 1816100000, 'BACS Payment'),
('BC', '', 'OPG BACS PAYMENT', 1816100000, 'BACS Payment'),
('CQ', '', 'CHEQUE PAYMENT', 1816100000, 'Cheque Payment'),
('RA', '', 'CREDIT REAPPLY', 1816100000, 'Reapply/Reallocate (money to invoice)'),
('BC', '', 'BACS TRANSFER', 1816100000, 'BACS Payment'),
('PC', '', 'CARD PAYMENT', 1816100000, 'Card Payment');

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