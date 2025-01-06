SET SEARCH_PATH TO supervision_finance;

INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null); -- main test client
INSERT INTO finance_client VALUES (99, 99, 'no entries', 'DEMANDED', null); -- empty client

-- CYPRESS TEST DATA: Create separate dataset for each test file

-- add-fee-reduction
INSERT INTO finance_client VALUES (2, 2, 'add-fee-reduction', 'DEMANDED', null);

-- add-manual-invoice
INSERT INTO finance_client VALUES (3, 3, 'add-manual-invoice', 'DEMANDED', null);

-- adjust-invoice
INSERT INTO finance_client VALUES (4, 4, 'adjust-invoice', 'DEMANDED', null);

INSERT INTO invoice VALUES (1, 4, 4, 'AD', 'AD11111/19', '2019-04-01', '2020-03-31', 10000, null, '2020-03-20', 10, '2020-03-16', null, null, null, '2019-06-06', 99); -- add credit

INSERT INTO invoice VALUES (2, 4, 4, 'S2', 'S203532/24', '2023-04-01', '2024-03-31', 32000, null, '2024-03-31', 10, '2024-03-31', null, null, null, '2024-03-31', 99); -- write off

INSERT INTO invoice VALUES (3, 4, 4, 'AD', 'AD03532/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', 99); -- add debit
INSERT INTO ledger VALUES (1, 'add-debit', '2024-04-11T08:36:40+00:00', '', 10000, '', 'CARD PAYMENT', 'CONFIRMED', 4, null, null, '11/04/2042', '12/04/2024', 1, '', '', 1, '05/05/2024', 65);
INSERT INTO ledger_allocation VALUES (1, 1, 3, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);

INSERT INTO invoice VALUES (4, 4, 4, 'AD', 'AD03533/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', '99'); -- write off reversal
INSERT INTO ledger VALUES (2, 'write-off', '2024-04-11T08:36:40+00:00', '', 10000, '', 'CREDIT WRITE OFF', 'APPROVED', 4);
INSERT INTO ledger_allocation VALUES (2, 2, 4, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, null, '2022-04-11');
INSERT INTO invoice_adjustment VALUES (1, 4, 4, '2022-12-04', 'CREDIT WRITE OFF', 10000, 'credit write off for 100.00', 'APPROVED', '2022-12-04T08:36:40+00:00', 65);

-- billing-history
INSERT INTO finance_client VALUES (5, 5, 'billing-history', 'DEMANDED', null);
INSERT INTO invoice VALUES (5, 5, 5, 'AD', 'AD44444/17', '2017-06-06', '2017-06-06', 10000, null, '2017-06-06', 10, '2017-06-06', null, null, null, '2017-06-06', 99);

-- cancel-fee-reduction
INSERT INTO finance_client VALUES (6, 6, 'cancel-fee-reduction', 'DEMANDED', null);
INSERT INTO fee_reduction VALUES (1, 6, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO fee_reduction VALUES (2, 6, 'REMISSION', null, '2020-04-01', '2021-03-31', 'notes', true, '2019-05-01', '2019-05-01', 1, '2019-05-01', 1, 'cancelled as duplicate');

-- customer-credit-balance
INSERT INTO finance_client VALUES (7, 7, 'customer-credit-balance', 'DEMANDED', null);
INSERT INTO invoice VALUES (6, 7, 7, 'AD', 'AD77777/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99); -- customer-credit-balance
INSERT INTO ledger VALUES (3, 'customer-credit-balance', '2024-04-11T08:36:40+00:00', '', 3000, '', 'CARD PAYMENT', 'CONFIRMED', 7, null, null, '11/04/2042', '12/04/2024', 1, '', '', 1, '05/05/2024', 65);
INSERT INTO ledger_allocation VALUES (3, 3, 6, '2024-04-11T08:36:40+00:00', 3000, 'ALLOCATED', null, 'customer-credit-balance', '2024-04-11', null);

-- fee-reductions
INSERT INTO finance_client VALUES (8, 8, 'fee-reductions', 'DEMANDED', null);
INSERT INTO fee_reduction VALUES (3, 8, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
INSERT INTO fee_reduction VALUES (4, 8, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-01-01')::DATE, CONCAT(date_part('year', now()), '-12-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);

-- invoices
INSERT INTO finance_client VALUES (9, 9, 'customer-credit-balance', 'DEMANDED', null);
INSERT INTO invoice VALUES (7, 9, 9, 'S2', 'S299999/19', '2019-03-16', '2019-03-16', 32000, null, '2019-03-16', 10, '2019-03-16', null, null, null, '2019-03-16', 99);
INSERT INTO ledger VALUES (4, 'invoice-test', '2024-04-11T08:36:40+00:00', '', 2000, '', 'CARD PAYMENT', 'CONFIRMED', 9, null, null, '11/04/2042', '12/04/2024', 1, '', '', 1, '05/05/2024', 65);
INSERT INTO ledger_allocation VALUES (4, 4, 7, '2024-04-11T08:36:40+00:00', 2000, 'ALLOCATED', null, 'invoices-test', '2024-04-11', null);
INSERT INTO invoice_adjustment VALUES (2, 9, 7, '2022-12-04', 'CREDIT MEMO', 1200, 'credit adjustment for 12.00', 'PENDING', '2022-12-04T08:36:40+00:00', 65);
INSERT INTO invoice_fee_range VALUES (1, 7, 'GENERAL', '2022-04-01', '2023-03-31', 32000);

-- pending-invoice-adjustments
INSERT INTO finance_client VALUES (10, 10, 'pending-invoice-adjustments', 'DEMANDED', null);
INSERT INTO invoice VALUES (8, 10, 10, 'AD', 'AD10101/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99);
INSERT INTO invoice_adjustment VALUES (3, 10, 8, '2022-04-11', 'CREDIT MEMO', 10000, 'credit adjustment for 100.00', 'PENDING', '2022-04-11T08:36:40+00:00', 65);

-- TEST CLIENT DATA: Add data for default client here

-- UPDATE SEQUENCES
SELECT setval('finance_client_id_seq', (SELECT MAX(id) FROM finance_client));
SELECT setval('fee_reduction_id_seq', (SELECT MAX(id) FROM fee_reduction));
SELECT setval('invoice_id_seq', (SELECT MAX(id) FROM invoice));
SELECT setval('invoice_adjustment_id_seq', (SELECT MAX(id) FROM invoice_adjustment));
SELECT setval('ledger_id_seq', (SELECT MAX(id) FROM ledger));
SELECT setval('ledger_allocation_id_seq', (SELECT MAX(id) FROM ledger_allocation));
SELECT setval('invoice_fee_range_id_seq', (SELECT MAX(id) FROM invoice_fee_range));
