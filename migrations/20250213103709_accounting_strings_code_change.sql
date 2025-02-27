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

INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description, line_description)
VALUES
    ('ZR', 'GA', 'CREDIT REMISSION', 4481102107, 'Remission Credit', 'GA fee reduction'),
    ('ZR', 'GS', 'CREDIT REMISSION', 4481102107, 'Remission Credit', 'GS fee reduction'),
    ('ZR', 'GT', 'CREDIT REMISSION', 4481102107, 'Remission Credit', 'GT fee reduction'),
    ('ZH', 'GA', 'CREDIT HARDSHIP', 4481102107, 'Hardship Credit', 'GA fee reduction'),
    ('ZH', 'GS', 'CREDIT HARDSHIP', 4481102107, 'Hardship Credit', 'GS fee reduction'),
    ('ZH', 'GT', 'CREDIT HARDSHIP', 4481102107, 'Hardship Credit', 'GT fee reduction'),
    ('ZE', 'GA', 'CREDIT EXEMPTION', 4481102108, 'Exemption Credit', 'GA fee reduction'),
    ('ZE', 'GS', 'CREDIT EXEMPTION', 4481102108, 'Exemption Credit', 'GS fee reduction'),
    ('ZE', 'GT', 'CREDIT EXEMPTION', 4481102108, 'Exemption Credit', 'GT fee reduction'),
    ('MCR', 'GA', 'CREDIT MEMO', 4481102107, 'Manual Credit', 'GA Manual credit'),
    ('MCR', 'GS', 'CREDIT MEMO', 4481102107, 'Manual Credit', 'GS Manual credit'),
    ('MCR', 'GT', 'CREDIT MEMO', 4481102107, 'Manual Credit', 'GT Manual credit'),
    ('MDR', 'GA', 'DEBIT MEMO', 4481102107, 'Manual Debit', 'GA Manual debit'),
    ('MDR', 'GS', 'DEBIT MEMO', 4481102107, 'Manual Credit', 'GS Manual debit'),
    ('MDR', 'GT', 'DEBIT MEMO', 4481102107, 'Manual Credit', 'GT Manual debit'),
    ('WO', 'GA', 'CREDIT WRITE OFF', 4481102107, 'Manual Write-off', 'GA Write-off'),
    ('WO', 'GS', 'CREDIT WRITE OFF', 4481102107, 'Manual Write-off', 'GS Write-off'),
    ('WO', 'GT', 'CREDIT WRITE OFF', 4481102107, 'Manual Write-off', 'GT Write-off'),
    ('WOR', 'GA', 'WRITE OFF REVERSAL', 4481102107, 'Write-off reversal', 'GA Write-off reversal'),
    ('WOR', 'GS', 'WRITE OFF REVERSAL', 4481102107, 'Write-off reversal', 'GS Write-off reversal'),
    ('WOR', 'GT', 'WRITE OFF REVERSAL', 4481102107, 'Write-off reversal', 'GT Write-off reversal');

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