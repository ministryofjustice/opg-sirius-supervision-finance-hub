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
('DD', 'AD', 'DIRECT DEBIT PAYMENT', 1816100000, 'Direct Debit Payment'),
('OC', 'AD', 'ONLINE CARD PAYMENT', 1816100000, 'Online Card Payment'),
('PC', 'AD', 'MOTO CARD PAYMENT', 1816100000, 'MOTO (phone) Card Payment'),
('BC', 'AD', 'SUPERVISION BACS PAYMENT', 1816100000, 'BACS Payment'),
('BC', 'AD', 'OPG BACS PAYMENT', 1816100000, 'BACS Payment'),
('CQ', 'AD', 'CHEQUE PAYMENT', 1816100000, 'Cheque Payment'),
('RA', 'AD', 'CREDIT REAPPLY', 1816100000, 'Reapply/Reallocate (money to invoice)'),
('', 'AD', 'BACS TRANSFER', 1816100000, 'BACS Payment'),
('PC', 'AD', 'CARD PAYMENT', 1816100000, 'Card Payment'),
('DD', 'GENERAL', 'DIRECT DEBIT PAYMENT', 1816100000, 'Direct Debit Payment'),
('OC', 'GENERAL', 'ONLINE CARD PAYMENT', 1816100000, 'Online Card Payment'),
('PC', 'GENERAL', 'MOTO CARD PAYMENT', 1816100000, 'MOTO (phone) Card Payment'),
('BC', 'GENERAL', 'SUPERVISION BACS PAYMENT', 1816100000, 'BACS Payment'),
('BC', 'GENERAL', 'OPG BACS PAYMENT', 1816100000, 'BACS Payment'),
('CQ', 'GENERAL', 'CHEQUE PAYMENT', 1816100000, 'Cheque Payment'),
('RA', 'GENERAL', 'CREDIT REAPPLY', 1816100000, 'Reapply/Reallocate (money to invoice)'),
('', 'GENERAL', 'BACS TRANSFER', 1816100000, 'BACS Transfer'),
('PC', 'GENERAL', 'CARD PAYMENT', 1816100000, 'Card Payment'),
('DD', 'MINIMAL', 'DIRECT DEBIT PAYMENT', 1816100000, 'Direct Debit Payment'),
('OC', 'MINIMAL', 'ONLINE CARD PAYMENT', 1816100000, 'Online Card Payment'),
('PC', 'MINIMAL', 'MOTO CARD PAYMENT', 1816100000, 'MOTO (phone) Card Payment'),
('BC', 'MINIMAL', 'SUPERVISION BACS PAYMENT', 1816100000, 'BACS Payment'),
('BC', 'MINIMAL', 'OPG BACS PAYMENT', 1816100000, 'BACS Payment'),
('CQ', 'MINIMAL', 'CHEQUE PAYMENT', 1816100000, 'Cheque Payment'),
('RA', 'MINIMAL', 'CREDIT REAPPLY', 1816100000, 'Reapply/Reallocate (money to invoice)'),
('', 'MINIMAL', 'BACS TRANSFER', 1816100000, 'BACS Payment'),
('PC', 'MINIMAL', 'CARD PAYMENT', 1816100000, 'Card Payment');

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