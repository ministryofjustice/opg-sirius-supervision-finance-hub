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