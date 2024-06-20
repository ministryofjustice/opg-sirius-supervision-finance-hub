SET SEARCH_PATH TO supervision_finance;

INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null);
INSERT INTO finance_client VALUES (2, 2, '1234', 'DEMANDED', null);
INSERT INTO finance_client VALUES (3, 3, '1234', 'DEMANDED', null); -- adjust-invoice.cy.js
ALTER SEQUENCE finance_client_id_seq RESTART WITH 4;

INSERT INTO fee_reduction VALUES (1, 1, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
ALTER SEQUENCE fee_reduction_id_seq RESTART WITH 2;

INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S206666/18', '2019-04-01', '2018-03-31', 32000, null, '2023-03-20', 10, '2018-03-16', null, null, null, '2018-06-06', 99);
INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'AD03531/19', '2019-04-01', '2020-03-31', 10000, null, '2020-03-20', 10, '2020-03-16', null, null, null, '2019-06-06', 99, 1);

INSERT INTO invoice VALUES (3, 3, 3, 'S2', 'S203532/24', '2023-04-01', '2024-03-31', 32000, null, '2024-03-31', 10, '2024-03-31', null, null, null, '2024-03-31', 99); -- add credit
INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'AD03532/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2024-03-31', null, null, null, '2024-03-31', 99); -- add debit
ALTER SEQUENCE invoice_id_seq RESTART WITH 5;

INSERT INTO ledger VALUES (1, 'adjustment123', '2022-04-11T08:36:40+00:00', '', 1200, 'credit adjustment for 12.00', 'CREDIT MEMO', 'PENDING', 1, 1, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (2, 'random1223', '2022-04-11T08:36:40+00:00', '', 18500, '', 'REMISSION', 'CONFIRMED', 1, 1, 1, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (3, 'addcredit', '2022-04-11T08:36:40+00:00', '', 10000, '', 'CREDIT MEMO', 'APPROVED', 3, null, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
ALTER SEQUENCE ledger_id_seq RESTART WITH 4;

INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 1200, 'PENDING', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (2, 2, 2, '2022-04-11T08:36:40+00:00', 8800, 'APPROVED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (3, 3, 4, '2022-04-11T08:36:40+00:00', 10000, 'APPROVED', null, 'Notes here', '2022-04-11', null);
ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 4;

INSERT INTO invoice_fee_range VALUES (1, 1, 'GENERAL', '2022-04-01', '2023-03-31', 32000);
INSERT INTO invoice_fee_range VALUES (2, 2, 'GENERAL', '2022-04-01', '2023-03-31', 10000);
ALTER SEQUENCE invoice_fee_range_id_seq RESTART WITH 3;