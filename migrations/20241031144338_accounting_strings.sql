-- +goose Up
CREATE TABLE cost_centre
(
    code                    INT     NOT NULL PRIMARY KEY,
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
    cost_centre              INTEGER REFERENCES cost_centre (code)
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
    id                INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    fee_type          VARCHAR NOT NULL,
    supervision_level VARCHAR NOT NULL,
    ledger_type       VARCHAR,
    account_code      BIGINT REFERENCES supervision_finance.account (code),
    description       VARCHAR
);

CREATE INDEX ON transaction_type (fee_type, supervision_level);
CREATE INDEX ON transaction_type (ledger_type, supervision_level);

INSERT INTO transaction_type (fee_type, supervision_level, ledger_type, account_code, description)
VALUES ('AD', 'AD', '', 4481102093, 'AD - Assessment deputy invoice'),
       ('S2', 'GENERAL', '', 4481102094, 'S2 - General invoice (Demanded)'),
       ('S3', 'MINIMAL', '', 4481102099, 'S3 - Minimal invoice (Demanded)'),
       ('B2', 'GENERAL', '', 4481102094, 'B2 - General invoice (Direct Debit)'),
       ('B3', 'MINIMAL', '', 4481102099, 'B3 - Minimal invoice (Direct Debit)'),
       ('SF', 'GENERAL', '', 4481102094, 'SF - Deceased invoice'),
       ('SF', 'MINIMAL', '', 4481102099, 'SF - Deceased invoice'),
       ('SE', 'GENERAL', '', 4481102094, 'SE - Order expired invoice'),
       ('SE', 'MINIMAL', '', 4481102099, 'SE - Order expired invoice'),
       ('SO', 'GENERAL', '', 4481102094, 'SO - Regained capacity invoice'),
       ('SO', 'MINIMAL', '', 4481102099, 'SO - Regained capacity invoice'),
       ('ZR', 'AD', 'CREDIT REMISSION', 4481102114, 'Remission Credit'),
       ('ZE', 'AD', 'CREDIT EXEMPTION', 4481102114, 'Exemption Credit'),
       ('ZH', 'AD', 'CREDIT HARDSHIP', 4481102114, 'Hardship Credit'),
       ('MCR', 'AD', 'CREDIT MEMO', 4481102093, 'Manual Credit'),
       ('MDR', 'AD', 'DEBIT MEMO', 4481102093, 'Manual Debit'),
       ('WO', 'AD', 'CREDIT WRITE OFF', 5356202100, 'Manual Write-off'),
       ('WOR', 'AD', 'WRITE OFF REVERSAL', 1816900000, 'Write-off reversal'),
       ('ZR', 'GENERAL', 'CREDIT REMISSION', 4481102115, 'Remission Credit'),
       ('ZE', 'GENERAL', 'CREDIT EXEMPTION', 4481102115, 'Exemption Credit'),
       ('ZH', 'GENERAL', 'CREDIT HARDSHIP', 4481102115, 'Hardship Credit'),
       ('MCR', 'GENERAL', 'CREDIT MEMO', 4481102094, 'Manual Credit'),
       ('MDR', 'GENERAL', 'DEBIT MEMO', 4481102094, 'Manual Debit'),
       ('WO', 'GENERAL', 'CREDIT WRITE OFF', 5356202102, 'Manual Write-off'),
       ('WOR', 'GENERAL', 'WRITE OFF REVERSAL', 1816900000, 'Write-off reversal'),
       ('ZR', 'MINIMAL', 'CREDIT REMISSION', 4481102120, 'Remission Credit'),
       ('ZE', 'MINIMAL', 'CREDIT EXEMPTION', 4481102120, 'Exemption Credit'),
       ('ZH', 'MINIMAL', 'CREDIT HARDSHIP', 4481102120, 'Hardship Credit'),
       ('MCR', 'MINIMAL', 'CREDIT MEMO', 4481102099, 'Manual Credit'),
       ('MDR', 'MINIMAL', 'DEBIT MEMO', 4481102099, 'Manual Debit'),
       ('WO', 'MINIMAL', 'CREDIT WRITE OFF', 5356202104, 'Manual Write-off'),
       ('WOR', 'MINIMAL', 'WRITE OFF REVERSAL', 1816900000, 'Write-off reversal'),
       ('GA', '', '', 4481102104, 'Guardianship assess invoice'),
       ('GS', '', '', 4481102105, 'Guardianship supervision invoice'),
       ('GT', '', '', 4481102106, 'Guardianship termination invoice'),
       ('UA', '', '', 1816100001, 'Unapply (money from invoice)'),
       ('OP', '', '', 1816100002, 'Overpayment'),
       ('CQR', '', '', 1816100005, 'Cheque Refund'),
       ('BCR', '', '', 1816100005, 'BACS Refund');

-- +goose Down

DROP TABLE transaction_type;
DROP TABLE account;
DROP TABLE cost_centre;
