-- +goose Up
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

INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description, line_description, is_receipt)
VALUES
    ('ZR', 'GA', 'CREDIT REMISSION', 4481102107, 'Remission Credit', 'GA fee reduction', false),
    ('ZR', 'GS', 'CREDIT REMISSION', 4481102107, 'Remission Credit', 'GS fee reduction', false),
    ('ZR', 'GT', 'CREDIT REMISSION', 4481102107, 'Remission Credit', 'GT fee reduction', false),
    ('ZH', 'GA', 'CREDIT HARDSHIP', 4481102107, 'Hardship Credit', 'GA fee reduction', false),
    ('ZH', 'GS', 'CREDIT HARDSHIP', 4481102107, 'Hardship Credit', 'GS fee reduction', false),
    ('ZH', 'GT', 'CREDIT HARDSHIP', 4481102107, 'Hardship Credit', 'GT fee reduction', false),
    ('ZE', 'GA', 'CREDIT EXEMPTION', 4481102108, 'Exemption Credit', 'GA fee reduction', false),
    ('ZE', 'GS', 'CREDIT EXEMPTION', 4481102108, 'Exemption Credit', 'GS fee reduction', false),
    ('ZE', 'GT', 'CREDIT EXEMPTION', 4481102108, 'Exemption Credit', 'GT fee reduction', false),
    ('MCR', 'GA', 'CREDIT MEMO', 4481102107, 'Manual Credit', 'GA Manual credit', false),
    ('MCR', 'GS', 'CREDIT MEMO', 4481102107, 'Manual Credit', 'GS Manual credit', false),
    ('MCR', 'GT', 'CREDIT MEMO', 4481102107, 'Manual Credit', 'GT Manual credit', false),
    ('MDR', 'GA', 'DEBIT MEMO', 4481102107, 'Manual Debit', 'GA Manual debit', false),
    ('MDR', 'GS', 'DEBIT MEMO', 4481102107, 'Manual Credit', 'GS Manual debit', false),
    ('MDR', 'GT', 'DEBIT MEMO', 4481102107, 'Manual Credit', 'GT Manual debit', false),
    ('WO', 'GA', 'CREDIT WRITE OFF', 4481102107, 'Manual Write-off', 'GA Write-off', false),
    ('WO', 'GS', 'CREDIT WRITE OFF', 4481102107, 'Manual Write-off', 'GS Write-off', false),
    ('WO', 'GT', 'CREDIT WRITE OFF', 4481102107, 'Manual Write-off', 'GT Write-off', false),
    ('WOR', 'GA', 'WRITE OFF REVERSAL', 4481102107, 'Write-off reversal', 'GA Write-off reversal', false),
    ('WOR', 'GS', 'WRITE OFF REVERSAL', 4481102107, 'Write-off reversal', 'GS Write-off reversal', false),
    ('WOR', 'GT', 'WRITE OFF REVERSAL', 4481102107, 'Write-off reversal', 'GT Write-off reversal', false);

-- +goose Down

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


DELETE FROM transaction_type WHERE supervision_level IN ('GA', 'GS', 'GT') AND fee_type IN ('ZR', 'ZH', 'ZE', 'MCR', 'MDR', 'WO', 'WOR');