-- +goose Up
INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description, line_description)
VALUES
    ('ZR', 'GA', '', 4481102107, 'Remission Credit', 'GA fee reduction'),
    ('ZR', 'GS', '', 4481102107, 'Remission Credit', 'GS fee reduction'),
    ('ZR', 'GT', '', 4481102107, 'Remission Credit', 'GT fee reduction'),
    ('ZH', 'GA', '', 4481102107, 'Hardship Credit', 'GA fee reduction'),
    ('ZH', 'GS', '', 4481102107, 'Hardship Credit', 'GS fee reduction'),
    ('ZH', 'GT', '', 4481102107, 'Hardship Credit', 'GT fee reduction'),
    ('ZE', 'GA', '', 4481102108, 'Exemption Credit', 'GA fee reduction'),
    ('ZE', 'GS', '', 4481102108, 'Exemption Credit', 'GS fee reduction'),
    ('ZE', 'GT', '', 4481102108, 'Exemption Credit', 'GT fee reduction'),
    ('MCR', 'GA', '', 4481102107, 'Manual Credit', 'GA Manual credit'),
    ('MCR', 'GS', '', 4481102107, 'Manual Credit', 'GS Manual credit'),
    ('MCR', 'GT', '', 4481102107, 'Manual Credit', 'GT Manual credit'),
    ('MDR', 'GA', '', 4481102107, 'Manual Debit', 'GA Manual debit'),
    ('MDR', 'GS', '', 4481102107, 'Manual Credit', 'GS Manual debit'),
    ('MDR', 'GT', '', 4481102107, 'Manual Credit', 'GT Manual debit'),
    ('WO', 'GA', '', 4481102107, 'Manual Write-off', 'GA Write-off'),
    ('WO', 'GS', '', 4481102107, 'Manual Write-off', 'GS Write-off'),
    ('WO', 'GT', '', 4481102107, 'Manual Write-off', 'GT Write-off'),
    ('WOR', 'GA', '', 4481102107, 'Write-off reversal', 'GA Write-off reversal'),
    ('WOR', 'GS', '', 4481102107, 'Write-off reversal', 'GS Write-off reversal'),
    ('WOR', 'GT', '', 4481102107, 'Write-off reversal', 'GT Write-off reversal');

-- +goose Down
DELETE FROM transaction_type WHERE supervision_level IN ('GA', 'GS', 'GT') AND fee_type IN ('ZR', 'ZH', 'ZE', 'MCR', 'MDR', 'WO', 'WOR');