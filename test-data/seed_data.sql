SET SEARCH_PATH TO supervision_finance;

INSERT INTO finance_client VALUES (1, 1, '1234', 'DEMANDED', null, 12300, 2222);
INSERT INTO finance_client VALUES (2, 2, '1234', 'DEMANDED', null, 0, 0);
INSERT INTO fee_reduction VALUES (1, 1, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
INSERT INTO invoice VALUES (1, 1, 1, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 12300, null, '2020-03-20',1, '2020-03-16', 10, null, 12300, '2019-06-06', 99);
