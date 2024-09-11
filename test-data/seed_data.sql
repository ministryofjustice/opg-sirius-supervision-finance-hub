SET SEARCH_PATH TO supervision_finance;

INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null);
INSERT INTO finance_client VALUES (2, 2, '1234', 'DEMANDED', null);
INSERT INTO finance_client VALUES (3, 3, 'adjust-invoice', 'DEMANDED', null);
INSERT INTO finance_client VALUES (4, 4, 'no entries', 'DEMANDED', null);
INSERT INTO finance_client VALUES (5, 5, 'add-fee-reduction', 'DEMANDED', null);
INSERT INTO finance_client VALUES (6, 6, 'customer-credit-balance', 'DEMANDED', null);
SELECT setval('finance_client_id_seq', (SELECT MAX(id) FROM finance_client));

INSERT INTO fee_reduction VALUES (1, 1, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
INSERT INTO fee_reduction VALUES (2, 1, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO fee_reduction VALUES (3, 2, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO fee_reduction VALUES (4, 2, 'REMISSION', null, '2020-04-01', '2021-03-31', 'notes', true, '2019-05-01', '2019-05-01', 1, '2019-05-01', 1, 'cancelled as duplicate');
SELECT setval('fee_reduction_id_seq', (SELECT MAX(id) FROM fee_reduction));

INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD04642/17', '2017-04-01', '2018-03-31', 10000, null, '2020-03-20', 10, '2018-03-16', null, null, null, '2017-06-06', 99);
INSERT INTO invoice VALUES (2, 1, 1, 'S2', 'S206666/18', '2018-04-01', '2019-03-31', 32000, null, '2023-03-20', 10, '2019-03-16', null, null, null, '2018-06-06', 99);
INSERT INTO invoice VALUES (3, 1, 1, 'AD', 'AD03531/19', '2019-04-01', '2020-03-31', 10000, null, '2020-03-20', 10, '2020-03-16', null, null, null, '2019-06-06', 99);
INSERT INTO invoice VALUES (4, 3, 3, 'S2', 'S203532/24', '2023-04-01', '2024-03-31', 32000, null, '2024-03-31', 10, '2024-03-31', null, null, null, '2024-03-31', 99); -- add credit
INSERT INTO invoice VALUES (5, 3, 3, 'AD', 'AD03532/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', 99); -- add debit
INSERT INTO invoice VALUES (6, 6, 6, 'AD', 'AD33333/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99); -- customer-credit-balance
SELECT setval('invoice_id_seq', (SELECT MAX(id) FROM invoice));

INSERT INTO invoice_adjustment VALUES (1, 1, 1, '2022-04-11', 'CREDIT MEMO', 1200, 'credit adjustment for 12.00', 'PENDING', '2022-04-11T08:36:40+00:00', 65);
INSERT INTO invoice_adjustment VALUES (2, 3, 3, '2022-04-11', 'CREDIT MEMO', 10000, 'credit adjustment for 100.00', 'PENDING', '2022-04-11T08:36:40+00:00', 65);
SELECT setval('invoice_adjustment_id_seq', (SELECT MAX(id) FROM invoice_adjustment));

INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T08:36:40+00:00', '', 18500, '', 'REMISSION', 'CONFIRMED', 1, 1, 1, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (2, 'paid', '2022-04-11T08:36:40+00:00', '', 10000, '', 'CARD PAYMENT', 'CONFIRMED', 1, null, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (3, 'unapply', '2022-04-11T08:36:40+00:00', '', 10000, '', 'REMISSION', 'CONFIRMED', 1, null, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (4, 'customer-credit-balance', '2024-04-11T08:36:40+00:00', '', 3000, '', 'CARD PAYMENT', 'CONFIRMED', 6, null, null, '11/04/2042', '12/04/2024', 1, '', '', 1, '05/05/2024', 65);
SELECT setval('ledger_id_seq', (SELECT MAX(id) FROM ledger));

INSERT INTO ledger_allocation VALUES (1, 2, 3, '2022-04-11T08:36:40+00:00', 8800, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (2, 3, 5, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (3, 2, 1, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (4, 3, 1, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (5, 3, 1, '2022-04-11T08:36:40+00:00', -10000, 'UNAPPLIED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (6, 4, 6, '2024-04-11T08:36:40+00:00', 3000, 'ALLOCATED', null, 'customer-credit-balance', '2024-04-11', null);
SELECT setval('ledger_allocation_id_seq', (SELECT MAX(id) FROM ledger_allocation));

INSERT INTO invoice_fee_range VALUES (1, 1, 'GENERAL', '2022-04-01', '2023-03-31', 32000);
INSERT INTO invoice_fee_range VALUES (2, 2, 'GENERAL', '2022-04-01', '2023-03-31', 10000);
INSERT INTO invoice_fee_range VALUES (3, 3, 'GENERAL', '2023-04-01', '2024-03-31', 10000);
SELECT setval('invoice_fee_range_id_seq', (SELECT MAX(id) FROM invoice_fee_range));
