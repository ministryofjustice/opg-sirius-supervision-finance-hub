-- +goose Up
SELECT 'up SQL query';
CREATE TABLE cost_centre
(
    code                    BIGINT  NOT NULL PRIMARY KEY,
    cost_centre_description VARCHAR NOT NULL
);

INSERT INTO cost_centre (code, cost_centre_description)
VALUES (99999999, 'BALANCE SHEET'),
       (10482009, 'Supervision Investigations'),
       (10486000, 'Allocations, HW & SIS BISD');

CREATE TABLE account
(
    code                     BIGINT  NOT NULL PRIMARY KEY,
    account_code_description VARCHAR NOT NULL,
    cost_centre              BIGINT REFERENCES cost_centre (code)
);

INSERT INTO account(code, account_code_description, cost_centre)
VALUES (4481102093, 'INC - RECEIPT OF FEES AND CHARGES - Appoint Deputy',
        10482009),                                                                                   -- AD fee invoices, manual credits, and debits
       (4481102094, 'INC - RECEIPT OF FEES AND CHARGES - Supervision Fee 1',
        10482009),                                                                                   --	General fee invoices, manual credits, and debits
       (4481102099, 'INC - RECEIPT OF FEES AND CHARGES - Annual Admin Fee 3',
        10482009),                                                                                   --	Minimal fee invoices, manual credits and debits
       (4481102114, 'INC - RECEIPT OF FEES AND CHARGES - Rem Appoint Deputy',
        10482009),                                                                                   --	Fee reduction adjustments to AD fees
       (4481102115, 'INC - RECEIPT OF FEES AND CHARGES - Rem Sup Fee 1',
        10482009),                                                                                   --	Fee reduction adjustments to General fees
       (4481102120, 'INC - RECEIPT OF FEES AND CHARGES - Rem Annual Admin Fee 3',
        10482009),                                                                                   --	Fee reduction adjustments to Minimal fees
       (5356202100, 'EXP - IMPAIRMENT - BAD DEBTS-Appoint Deputy Write Off',
        10482009),                                                                                   --	Write-off adjustments to AD fees
       (5356202102, 'EXP - IMPAIRMENT - BAD DEBTS-Sup Fee 2 Write Off	Write-off',
        10482009),                                                                                   -- adjustments to General fees
       (5356202104, 'EXP - IMPAIRMENT - BAD DEBTS-Sup Fee 3 Write Off	Write-off',
        10482009),                                                                                   -- adjustments to Minimal fees
       (1816100000, 'CA - TRADE RECEIVABLES', 99999999),                                             --	Invoice adjustments on receivables
       (1816100001, 'CA - TRADE RECEIVABLES - UNAPPLIED RECEIPTS',
        99999999),                                                                                   --	Adjustments to unapplied monies on receivables
       (1816100002, 'CA - TRADE RECEIVABLES - ON ACCOUNT RECEIPTS',
        99999999),                                                                                   --	Adjustments to on account receipts (overpayments)
       (1816100005, 'CA - TRADE RECEIVABLES - AR REFUNDS CTRL', 99999999),                           --	Refund adjustments to receivables
       (1816900000, 'CA - OTHER RECEIVABLES', 99999999),                                             --	Write-off reversals
       (1841102050, 'CA - CASH BALANCES HELD WITH THE GBS - RBS PUBLIC GUARDIAN ***2583', 99999999), --	OPG bank account
       (1841102088, 'CA - CASH BALANCES HELD WITH THE GBS - OPG SUPERVISION BACS BANK ACCOUNT',
        99999999),                                                                                   --	Supervision bank account
       (4481102104, 'INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP ASSESS', 10486000),            --	Guardiansip assess
       (4481102105, 'INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP SUPERVISION FEE',
        10486000),                                                                                   --	Guardianship supervision fee
       (4481102106, 'INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP TERMINATED', 10486000),        --	Guardianship terminated
       (4481102107, 'INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP FEE REMISSION',
        10486000),                                                                                   --	Guardianship fee remission
       (4481102108, 'INC - RECEIPT OF FEES AND CHARGES - GUARDIANSHIP FEE EXEMPTION',
        10486000); --	Guardianship fee exemption

CREATE TABLE transaction_type
(
    id           INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    type_id      VARCHAR NOT NULL,
    fee_type     VARCHAR NOT NULL,
    account_code BIGINT REFERENCES account (code),
    description  VARCHAR
);


INSERT INTO transaction_type (type_id, fee_type, account_code, description)
VALUES ('AD', 'AD', 4481102093, 'AD - Assessment deputy invoice'),
       ('S2', 'GENERAL', 4481102094, 'S2 - General invoice (Demanded)'),
       ('S3', 'MINIMAL', 4481102099, 'S3 - Minimal invoice (Demanded)'),
       ('B2', 'GENERAL', 4481102094, 'B2 - General invoice (Direct Debit)'),
       ('B3', 'MINIMAL', 4481102099, 'B3 - Minimal invoice (Direct Debit)'),
       ('SF', 'GENERAL', 4481102094, 'SF - Deceased invoice'),
       ('SF', 'MINIMAL', 4481102099, 'SF - Deceased invoice'),
       ('SE', 'GENERAL', 4481102094, 'SE - Order expired invoice'),
       ('SE', 'MINIMAL', 4481102099, 'SE - Order expired invoice'),
       ('SO', 'GENERAL', 4481102094, 'SO - Regained capacity invoice'),
       ('SO', 'MINIMAL', 4481102099, 'SO - Regained capacity invoice'),
       ('GA', '', 4481102104, 'Guardianship assess invoice'),
       ('GS', '', 4481102105, 'Guardianship supervision invoice'),
       ('GT', '', 4481102106, 'Guardianship termination invoice'),
       ('ZR', 'AD', 4481102114, 'Remission Credit'),
       ('ZE', 'AD', 4481102114, 'Exemption Credit'),
       ('ZH', 'AD', 4481102114, 'Hardship Credit'),
       ('MCR', 'AD', 4481102093, 'Manual Credit'),
       ('MDR', 'AD', 4481102093, 'Manual Debit'),
       ('WO', 'AD', 5356202100, 'Manual Write-off'),
       ('ZR', 'GENERAL', 4481102115, 'Remission Credit'),
       ('ZE', 'GENERAL', 4481102115, 'Exemption Credit'),
       ('ZH', 'GENERAL', 4481102115, 'Hardship Credit'),
       ('MCR', 'GENERAL', 4481102094, 'Manual Credit'),
       ('MDR', 'GENERAL', 4481102094, 'Manual Debit'),
       ('WO', 'GENERAL', 5356202102, 'Manual Write-off'),
       ('ZR', 'MINIMAL', 4481102120, 'Remission Credit'),
       ('ZE', 'MINIMAL', 4481102120, 'Exemption Credit'),
       ('ZH', 'MINIMAL', 4481102120, 'Hardship Credit'),
       ('MCR', 'MINIMAL', 4481102099, 'Manual Credit'),
       ('MDR', 'MINIMAL', 4481102099, 'Manual Debit'),
       ('WO', 'MINIMAL', 5356202104, 'Manual Write-off'),
       ('WOR', '', 1816900000, 'Write-off reversal'),
       ('UA', '', 1816100001, 'Unapply (money from invoice)'),
       ('OP', '', 1816100002, 'Overpayment'),
       ('CQR', '', 1816100005, 'Cheque Refund'),
       ('BCR', '', 1816100005, 'BACS Refund');
-- not sure on these ones
--       ('DD', '', 1, 'Direct Debit Payment'),
--       ('OC', '', 1, 'Online Card Payment'),
--       ('PC', '', 1, 'MOTO (phone) Card Payment'),
--       ('BC', '', 1, 'BACS Payment'),
--       ('CQ', '', 1, 'Cheque Payment'),
--       ('CR', '', 1, 'Cheque reversal'),
--       ('RA', '', 1, 'Reapply/Reallocate (money to invoice)')

-- +goose Down
SELECT 'down SQL query';
DROP TABLE cost_centre;
DROP TABLE account;
DROP TABLE transaction_type;
