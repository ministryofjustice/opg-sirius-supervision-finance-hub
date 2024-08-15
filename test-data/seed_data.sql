SET SEARCH_PATH TO supervision_finance;

INSERT INTO finance_client VALUES (nextval('finance_client_id_seq'), 1, '1234', 'DEMANDED', null);
INSERT INTO finance_client VALUES (nextval('finance_client_id_seq'), 2, '1234', 'DEMANDED', null);
INSERT INTO finance_client VALUES (nextval('finance_client_id_seq'), 3, 'adjust-invoice', 'DEMANDED', null);
INSERT INTO finance_client VALUES (nextval('finance_client_id_seq'), 4, 'no entries', 'DEMANDED', null);
INSERT INTO finance_client VALUES (nextval('finance_client_id_seq'), 5, 'add-fee-reduction', 'DEMANDED', null);
INSERT INTO finance_client VALUES (nextval('finance_client_id_seq'), 6, 'customer-credit-balance', 'DEMANDED', null);

INSERT INTO fee_reduction VALUES (nextval('fee_reduction_id_seq'), 1, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
INSERT INTO fee_reduction VALUES (nextval('fee_reduction_id_seq'), 1, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO fee_reduction VALUES (nextval('fee_reduction_id_seq'), 2, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO fee_reduction VALUES (nextval('fee_reduction_id_seq'), 2, 'REMISSION', null, '2020-04-01', '2021-03-31', 'notes', true, '2019-05-01', '2019-05-01', 1, '2019-05-01', 1, 'cancelled as duplicate');

INSERT INTO invoice VALUES (nextval('invoice_id_seq'), 1, 1, 'AD', 'AD04642/17', '2017-04-01', '2018-03-31', 10000, null, '2020-03-20', 10, '2018-03-16', null, null, null, '2017-06-06', 99);
INSERT INTO invoice VALUES (nextval('invoice_id_seq'), 1, 1, 'S2', 'S206666/18', '2018-04-01', '2019-03-31', 32000, null, '2023-03-20', 10, '2019-03-16', null, null, null, '2018-06-06', 99);
INSERT INTO invoice VALUES (nextval('invoice_id_seq'), 1, 1, 'AD', 'AD03531/19', '2019-04-01', '2020-03-31', 10000, null, '2020-03-20', 10, '2020-03-16', null, null, null, '2019-06-06', 99);
INSERT INTO invoice VALUES (nextval('invoice_id_seq'), 3, 3, 'S2', 'S203532/24', '2023-04-01', '2024-03-31', 32000, null, '2024-03-31', 10, '2024-03-31', null, null, null, '2024-03-31', 99); -- add credit
INSERT INTO invoice VALUES (nextval('invoice_id_seq'), 3, 3, 'AD', 'AD03532/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', 99); -- add debit
INSERT INTO invoice VALUES (nextval('invoice_id_seq'), 6, 6, 'AD', 'AD33333/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99); -- customer-credit-balance

INSERT INTO ledger VALUES (nextval('ledger_id_seq'), 'adjustment123', '2022-04-11T08:36:40+00:00', '', 1200, 'credit adjustment for 12.00', 'CREDIT MEMO', 'PENDING', 1, 1, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (nextval('ledger_id_seq'), 'random1223', '2022-04-11T08:36:40+00:00', '', 18500, '', 'REMISSION', 'CONFIRMED', 1, 1, 1, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (nextval('ledger_id_seq'), 'addcredit', '2022-04-11T08:36:40+00:00', '', 10000, 'credit adjustment for 100.00', 'CREDIT MEMO', 'PENDING', 3, null, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (nextval('ledger_id_seq'), 'paid', '2022-04-11T08:36:40+00:00', '', 10000, '', 'CARD PAYMENT', 'CONFIRMED', 1, null, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (nextval('ledger_id_seq'), 'unapply', '2022-04-11T08:36:40+00:00', '', 10000, '', 'REMISSION', 'CONFIRMED', 1, null, null, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);
INSERT INTO ledger VALUES (nextval('ledger_id_seq'), 'customer-credit-balance', '2024-04-11T08:36:40+00:00', '', 3000, '', 'CARD PAYMENT', 'CONFIRMED', 6, null, null, '11/04/2042', '12/04/2024', 1, '', '', 1, '05/05/2024', 65);

INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 1, 2, '2022-04-11T08:36:40+00:00', 1200, 'PENDING', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 2, 3, '2022-04-11T08:36:40+00:00', 8800, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 3, 5, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 4, 1, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 5, 1, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 5, 1, '2022-04-11T08:36:40+00:00', -10000, 'UNAPPLIED', null, 'Notes here', '2022-04-11', null);
INSERT INTO ledger_allocation VALUES (nextval('ledger_allocation_id_seq'), 6, 6, '2024-04-11T08:36:40+00:00', 3000, 'ALLOCATED', null, 'customer-credit-balance', '2024-04-11', null);

INSERT INTO invoice_fee_range VALUES (nextval('invoice_fee_range_id_seq'), 1, 'GENERAL', '2022-04-01', '2023-03-31', 32000);
INSERT INTO invoice_fee_range VALUES (nextval('invoice_fee_range_id_seq'), 2, 'GENERAL', '2022-04-01', '2023-03-31', 10000);
INSERT INTO invoice_fee_range VALUES (nextval('invoice_fee_range_id_seq'), 3, 'GENERAL', '2023-04-01', '2024-03-31', 10000);
