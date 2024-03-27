INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null, 12300, 2222);
INSERT INTO finance_client VALUES (2, 2, '1234', 'DEMANDED', null, 0, 0);
INSERT INTO fee_reduction VALUES (1, 1, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 12300, null, 1, '2020-03-20', 10, '2020-03-16', null, null, 12300, '2019-06-06', 99);
INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T08:36:40+00:00', '', 12300, '', 'Unknown Credit', 'Confirmed', 1, 1, 1, '11/04/2022', '11/04/2022', 1254, '', 1);
INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 12300, 'Confirmed', null, 'Notes here', '2022-04-11', null)
