SET SEARCH_PATH TO supervision_finance;

INSERT INTO finance_client VALUES (1001, 1, '1234', 'DEMANDED', null, '11111111'); -- main test client
INSERT INTO finance_client VALUES (99001, 99, 'no entries', 'DEMANDED', null, '00000000'); -- empty client

-- CYPRESS TEST DATA: Create separate dataset for each test file

-- add-fee-reduction
INSERT INTO finance_client VALUES (2001, 2, 'add-fee-reduction', 'DEMANDED', null, '22222222');

-- add-manual-invoice
INSERT INTO finance_client VALUES (3001, 3, 'add-manual-invoice', 'DEMANDED', null, '33333333');

-- adjust-invoice
INSERT INTO finance_client VALUES (4001, 4, 'adjust-invoice', 'DEMANDED', null, '44444444');
INSERT INTO invoice VALUES (1, 4, 4001, 'AD', 'AD11111/19', '2019-04-01', '2020-03-31', 10000, null, '2020-03-20', 10, '2020-03-16', null, null, null, '2019-06-06', 99); -- add credit
INSERT INTO invoice VALUES (2, 4, 4001, 'S2', 'S203532/24', '2023-04-01', '2024-03-31', 32000, null, '2024-03-31', 10, '2024-03-31', null, null, null, '2024-03-31', 99); -- write off

INSERT INTO invoice VALUES (3, 4, 4001, 'AD', 'AD03532/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', 99); -- add debit
INSERT INTO ledger VALUES (1, 'add-debit', '2024-04-11T08:36:40+00:00', '', 10000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 4001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (1, 1, 3, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'Notes here', '2022-04-11', null);

INSERT INTO invoice VALUES (4, 4, 4001, 'AD', 'AD03533/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', '99'); -- write off reversal
INSERT INTO ledger VALUES (2, 'write-off', '2024-04-11T08:36:40+00:00', '', 10000, '', 'CREDIT WRITE OFF', 'CONFIRMED', 4001);
INSERT INTO ledger_allocation VALUES (2, 2, 4, '2022-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, null, '2022-04-11');
INSERT INTO invoice_adjustment VALUES (1, 4001, 4, '2022-12-04', 'CREDIT WRITE OFF', 10000, 'credit write off for 100.00', 'APPROVED', '2022-12-04T08:36:40+00:00', 2);

INSERT INTO invoice VALUES (5, 4, 4001, 'AD', 'AD03534/24', '2023-04-01', '2024-03-31', 10000, null, '2024-03-31', 10, '2023-04-01', null, null, null, '2024-03-31', '99'); -- fee reduction reversal
INSERT INTO ledger VALUES (3, 'write-off-fr', '2024-04-11T08:36:40+00:00', '', 5000, '', 'CREDIT MEMO', 'CONFIRMED', 4001);
INSERT INTO ledger_allocation VALUES (3, 3, 5, '2022-04-11T08:36:40+00:00', 5000, 'ALLOCATED', null, null, '2022-04-11');
INSERT INTO fee_reduction VALUES (1, 4001, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO ledger VALUES (4, 'fee-reduction', '2024-04-11T08:36:40+00:00', '', 5000, '', 'CREDIT REMISSION', 'CONFIRMED', 4001, null, 1);
INSERT INTO ledger_allocation VALUES (4, 4, 5, '2022-04-11T08:36:40+00:00', 5000, 'ALLOCATED', null, null, '2022-04-11');

-- billing-history
INSERT INTO finance_client VALUES (5001, 5, 'billing-history', 'DEMANDED', null, '55555555');
INSERT INTO invoice VALUES (6, 5, 5001, 'AD', 'AD44444/17', '2017-06-06', '2017-06-06', 10000, null, '2017-06-06', 10, '2017-06-06', null, null, null, '2017-06-06', 99);

-- cancel-fee-reduction
INSERT INTO finance_client VALUES (6001, 6, 'cancel-fee-reduction', 'DEMANDED', null, '66666666');
INSERT INTO fee_reduction VALUES (2, 6001, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-04-01')::DATE, CONCAT(date_part('year', now()), '-03-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);
INSERT INTO fee_reduction VALUES (3, 6001, 'REMISSION', null, '2020-04-01', '2021-03-31', 'notes', true, '2019-05-01', '2019-05-01', 1, '2019-05-01', 1, 'cancelled as duplicate');

-- customer-credit-balance
INSERT INTO finance_client VALUES (7001, 7, 'customer-credit-balance', 'DEMANDED', null, '77777777');
INSERT INTO invoice VALUES (7, 7, 7001, 'AD', 'AD77777/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99); -- customer-credit-balance
INSERT INTO ledger VALUES (5, 'customer-credit-balance', '2024-04-11T08:36:40+00:00', '', 3000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 7001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (5, 5, 7, '2024-04-11T08:36:40+00:00', 3000, 'ALLOCATED', null, 'customer-credit-balance', '2024-04-11', null);

-- fee-reductions
INSERT INTO finance_client VALUES (8001, 8, 'fee-reductions', 'DEMANDED', null, '88888888');
INSERT INTO fee_reduction VALUES (4, 8001, 'REMISSION', null, '2019-04-01', '2020-03-31', 'notes', false, '2019-05-01');
INSERT INTO fee_reduction VALUES (5, 8001, 'HARDSHIP', null, CONCAT(date_part('year', now()), '-01-01')::DATE, CONCAT(date_part('year', now()), '-12-31')::DATE + INTERVAL '1 year', 'current reduction', false, '2020-05-01', '2020-05-01', 1);

-- invoices
INSERT INTO finance_client VALUES (9001, 9, 'customer-credit-balance', 'DEMANDED', null, '99999999');
INSERT INTO invoice VALUES (8, 9, 9001, 'S2', 'S299999/19', '2019-03-16', '2019-03-16', 32000, null, '2019-03-16', 10, '2019-03-16', null, null, null, '2019-03-16', 99);
INSERT INTO ledger VALUES (6, 'invoice-test', '2024-04-11T08:36:40+00:00', '', 2000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 9001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (6, 6, 8, '2024-04-11T08:36:40+00:00', 2000, 'ALLOCATED', null, 'invoices-test', '2024-04-11', null);
INSERT INTO invoice_adjustment VALUES (2, 9001, 8, '2022-12-04', 'CREDIT MEMO', 1200, 'credit adjustment for 12.00', 'PENDING', '2022-12-04T08:36:40+00:00', 2);
INSERT INTO invoice_fee_range VALUES (1, 8, 'GENERAL', '2022-04-01', '2023-03-31', 32000);
-- this transaction should be ignored as the ledger contains a legacy status
INSERT INTO ledger VALUES (7, 'ignore-me', '2024-04-11T08:36:40+00:00', '', 2000, '', 'ONLINE CARD PAYMENT', 'APPROVED', 9001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (7, 7, 8, '2024-04-11T08:36:40+00:00', 2000, 'ALLOCATED', null, 'invoices-test', '2024-04-11', null);

-- invoice-adjustments
INSERT INTO finance_client VALUES (10001, 10, 'invoice-adjustments', 'DEMANDED', null, '10101010');
INSERT INTO invoice VALUES (9, 10, 10001, 'AD', 'AD10101/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99);
INSERT INTO invoice_adjustment VALUES (3, 10001, 9, '2022-04-11', 'CREDIT MEMO', 10000, 'credit adjustment for 100.00', 'PENDING', '2022-04-11T08:36:40+00:00', 4);

-- permissions
INSERT INTO finance_client VALUES (11001, 11, 'permissions', 'DEMANDED', null, '11011011');
INSERT INTO invoice VALUES (10, 11, 11001, 'AD', 'AD10102/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99);
INSERT INTO invoice_adjustment VALUES (4, 11001, 10, '2022-04-11', 'CREDIT MEMO', 10000, 'credit adjustment for 100.00', 'PENDING', '2022-04-11T08:36:40+00:00', 2);
INSERT INTO fee_reduction VALUES (6, 11001, 'REMISSION', null, CONCAT(date_part('year', now()), '-01-01')::DATE, CONCAT(date_part('year', now()), '-12-31')::DATE + INTERVAL '1 year', 'notes', false, '2019-05-01');

-- payments - events
INSERT INTO finance_client VALUES (12001, 12, 'payments', 'DEMANDED', null, '12121212');
INSERT INTO invoice VALUES (11, 12, 12001, 'AD', 'AD12121/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99);

-- payments - api
INSERT INTO finance_client VALUES (13001, 13, 'paymentsapi', 'DEMANDED', null, '13131313');
INSERT INTO invoice VALUES (12, 13, 13001, 'AD', 'AD33333/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2025-03-31', 99);

-- refunds
INSERT INTO finance_client VALUES (14001, 14, 'refunds', 'DEMANDED', null, '14141414');
INSERT INTO ledger VALUES (8, 'refund-credit', '2024-04-11T08:36:40+00:00', '', 13000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 14001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (8, 8, NULL, '2024-04-11T08:36:40+00:00', -13000, 'UNAPPLIED', null, 'refund-credit', '2024-04-11', null);
INSERT INTO refund VALUES (1, 14001, '2025-06-01', 12340, 'APPROVED', 'Fulfilled refund', 1, '2025-06-01 00:00:00', 2, '2025-06-02 00:00:00', '2025-06-03 00:00:00', NULL, '2025-06-06 00:00:00');
INSERT INTO refund VALUES (2, 14001, '2024-05-01', 12341, 'PENDING', 'Pending refund', 1, '2024-05-01 00:00:00', NULL, NULL);
INSERT INTO refund VALUES (3, 14001, '2023-04-01', 12342, 'APPROVED', 'Approved refund', 1, '2023-04-02 00:00:00', 2, '2023-04-06 00:00:00');
INSERT INTO refund VALUES (4, 14001, '2022-03-01', 12343, 'REJECTED', 'Rejected refund', 1, '2022-03-02 00:00:00', 2, '2022-03-06 00:00:00');
INSERT INTO refund VALUES (5, 14001, '2021-02-01', 12344, 'APPROVED', 'Processing refund', 1, '2021-02-02 00:00:00', 2, '2021-02-03 00:00:00', '2021-02-06 00:00:00');
INSERT INTO refund VALUES (6, 14001, '2020-01-01', 12345, 'APPROVED', 'Cancelled refund', 1, '2020-01-02 00:00:00', 2, '2020-01-03 00:00:00', '2020-01-04 00:00:00', '2020-01-06 00:00:00', null, 2);
INSERT INTO bank_details VALUES (1, 2, 'Reginald Refund', '12345678', '11-22-33');
INSERT INTO bank_details VALUES (2, 3, 'Reginald Refund', '12345678', '11-22-33');

-- add refunds
INSERT INTO finance_client VALUES (15001, 15, 'addrefunds', 'DEMANDED', null, '15151515');
INSERT INTO invoice VALUES (13, 15, 15001, 'AD', 'AD15151/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);
INSERT INTO ledger VALUES (9, 'add-refund-credit', '2024-04-11T08:36:40+00:00', '', 13000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 15001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (10, 9, 13, '2024-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'refund-credit', '2024-04-11', null);
INSERT INTO ledger_allocation VALUES (11, 9, 13, '2024-04-11T08:36:40+00:00', -3000, 'UNAPPLIED', null, 'refund-credit', '2024-04-11', null);

-- add refunds - no credit
INSERT INTO finance_client VALUES (16001, 16, 'addrefundsnocredit', 'DEMANDED', null, '16161616');
INSERT INTO invoice VALUES (14, 16, 16001, 'AD', 'AD16161/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);
INSERT INTO ledger VALUES (10, 'add-refund-no-credit', '2024-04-11T08:36:40+00:00', '', 13000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 16001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (12, 10, 14, '2024-04-11T08:36:40+00:00', 13000, 'ALLOCATED', null, 'refund-credit', '2024-04-11', null);

-- refund decisions
INSERT INTO finance_client VALUES (17001, 17, 'refunddecision', 'DEMANDED', null, '17171717');
INSERT INTO ledger VALUES (11, 'refunddecision', '2024-04-11T08:36:40+00:00', '', 13000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 17001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-05', 2);
INSERT INTO ledger_allocation VALUES (13, 11, NULL, '2024-04-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'refund-credit', '2024-04-11', null);
INSERT INTO ledger_allocation VALUES (14, 11, NULL, '2024-04-11T08:36:40+00:00', -3000, 'UNAPPLIED', null, 'refund-credit', '2024-04-11', null);

INSERT INTO refund VALUES (7, 17001, '2024-06-02', 12341, 'PENDING', 'Approve me', 1, '2025-06-01 00:00:00', NULL, NULL);
INSERT INTO refund VALUES (8, 17001, '2024-06-01', 12341, 'PENDING', 'Reject me', 1, '2025-06-01 00:00:00', NULL, NULL);
INSERT INTO bank_details VALUES (3, 7, 'Donny Decisions', '12345678', '11-22-33');
INSERT INTO bank_details VALUES (4, 8, 'Donny Decisions', '12345678', '11-22-33');

-- cancel a processed refund
INSERT INTO finance_client VALUES (18001, 18, 'cancelrefund', 'DEMANDED', null, '18181818');
INSERT INTO refund VALUES (9, 18001, '2021-06-01', 12344, 'APPROVED', 'Cancel me 1', 1, '2022-06-01 00:00:00', 1, '2022-06-06 00:00:00', '2025-06-06 00:00:00');

-- cancel direct debit
INSERT INTO finance_client VALUES (19001, 19, 'canceldirectdebit', 'DIRECT DEBIT', null, '19191919');
INSERT INTO pending_collection VALUES (1, 19001, now()::DATE - 1, 10000, 'COLLECTED', NULL, '2025-06-01 00:00:00', 1);
INSERT INTO pending_collection VALUES (2, 19001, now()::DATE + 1, 10000, 'PENDING', NULL, '2025-06-01 00:00:00', 1);
INSERT INTO pending_collection VALUES (3, 19001, now()::DATE + 10, 10000, 'PENDING', NULL, '2025-06-01 00:00:00', 1);

-- billing history payments
INSERT INTO finance_client VALUES (20001, 20, 'paymentevents', 'DEMANDED', null, '20202020');
INSERT INTO invoice VALUES (15, 20, 20001, 'AD', 'AD16162/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 2);
INSERT INTO invoice VALUES (16, 20, 20001, 'AD', 'AD16163/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-05-12T12:01:59+00:00', 2);
INSERT INTO ledger VALUES (12, 'moto payment', '2024-05-11T08:36:40+00:00', '', 13000, '', 'MOTO CARD PAYMENT', 'CONFIRMED', 20001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-11 12:01:58', 2);
INSERT INTO ledger_allocation VALUES (15, 12, 15, '2024-05-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'moto-payment', '2024-05-11', null);
INSERT INTO ledger_allocation VALUES (16, 12, NULL, '2024-05-11T08:36:40+00:00', -3000, 'UNAPPLIED', null, 'moto-payment', '2024-05-11', null);
INSERT INTO ledger VALUES (13, 'reapply', '2024-05-12T12:02:41+00:00', '', 3000, '', 'CREDIT REAPPLY', 'CONFIRMED', 20001, null, null, '2024-04-12', '2024-04-12', 1, '', '', 1, '2024-05-12 12:02:00', 2);
INSERT INTO ledger_allocation VALUES (17, 13, 16, '2024-05-12T08:36:41+00:00', 3000, 'REAPPLIED', null, 'moto-payment', '2024-05-12', null);
INSERT INTO ledger VALUES (14, 'moto payment 2', '2024-05-13T08:39:40+00:00', '', 7000, '', 'MOTO CARD PAYMENT', 'CONFIRMED', 20001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-13 12:05:34', 2);
INSERT INTO ledger_allocation VALUES (18, 14, 15, '2024-05-13T08:39:40+00:00', 7000, 'ALLOCATED', null, 'moto-payment', '2024-05-13', null);
INSERT INTO ledger VALUES (15, 'moto payment reversal', '2024-05-14T08:36:40+00:00', '', -7000, '', 'MOTO CARD PAYMENT', 'CONFIRMED', 20001, null, null, '2024-04-11', '2024-04-12', 1, '', '', 1, '2024-05-14 12:05:34', 2);
INSERT INTO ledger_allocation VALUES (19, 15, 15, '2024-05-14T08:36:40+00:00', -7000, 'ALLOCATED', null, 'moto-payment', '2024-05-14', null);

-- create ledger for pending collection (event)
INSERT INTO finance_client VALUES (21001, 21, 'pendingcollection', 'DIRECT DEBIT', null, '21212100');
INSERT INTO invoice VALUES (17, 21, 21001, 'AD', 'AD212100/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);
INSERT INTO pending_collection VALUES (4, 21001, '2025-08-01', 10000, 'PENDING', NULL, '2025-08-21', 1);

-- create direct debit mandate/schedule
INSERT INTO finance_client VALUES (22001, 22, 'createdirectdebit', 'DEMANDED', null, '22222200');
INSERT INTO invoice VALUES (18, 22, 22001, 'AD', 'AD222200/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);

-- failed direct debit allpay validation
INSERT INTO finance_client VALUES (23001, 23, 'allpayvalidation', 'DIRECT DEBIT', null, '23232300');
INSERT INTO invoice VALUES (19, 23, 23001, 'AD', 'AD232300/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);
INSERT INTO finance_client VALUES (23003, 25, 'allpayvalidation', 'DIRECT DEBIT', null, '23232301');
INSERT INTO invoice VALUES (20, 25, 23003, 'AD', 'AD2323003/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);
INSERT INTO finance_client VALUES (23004, 26, 'allpayvalidation', 'DIRECT DEBIT', null, '23232302');
INSERT INTO invoice VALUES (21, 26, 23004, 'AD', 'AD2323004/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-04-10T08:36:40+00:00', 99);

-- cancel an approved refund
INSERT INTO finance_client VALUES (24001, 24, 'cancelapprovedrefund', 'DEMANDED', null, '24242400');
INSERT INTO refund VALUES (12, 24001, '2021-06-01', 11122, 'APPROVED', 'Cancel me 2', 2, '2022-06-01 00:00:00', 1, '2022-06-06 00:00:00');

-- allpay fails
INSERT INTO finance_client VALUES (25001, 25, 'allpay_mandate_fail', 'DIRECT DEBIT', null, '25252500');
INSERT INTO finance_client VALUES (26001, 26, 'allpay_schedule_fail', 'DIRECT DEBIT', null, '26262600');
INSERT INTO finance_client VALUES (27001, 27, 'allpay_schedule_validation', 'DIRECT DEBIT', null, '27272700');

-- failed direct debit payments (api)
INSERT INTO finance_client VALUES (28001, 28, 'failedddpayment', 'DIRECT DEBIT', null, '28282800');
INSERT INTO invoice VALUES (22, 28, 28001, 'AD', 'AD282828/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-05-11T12:01:59+00:00', 2);
INSERT INTO ledger VALUES (16, 'dd payment 1', '2025-09-30T08:36:40+00:00', '', 6000, '', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 28001, null, null, '2025-09-30', '2025-09-30', 1, '', '', 1, '2025-09-30 12:01:58', 2);
INSERT INTO ledger_allocation VALUES (20, 16, 22, '2025-09-30T08:36:40+00:00', 6000, 'ALLOCATED', null, '', '22025-09-30', null);
INSERT INTO ledger VALUES (17, 'dd payment 2', '2025-10-08T08:36:40+00:00', '', 4000, '', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 28001, null, null, '2025-10-08', '2025-10-08', 1, '', '', 1, '2025-10-08 12:01:58', 2);
INSERT INTO ledger_allocation VALUES (21, 17, 22, '2025-10-08T08:36:40+00:00', 4000, 'ALLOCATED', null, '', '2025-10-08', null);

-- allpay e2e
INSERT INTO finance_client VALUES (29001, 29, 'allpaye2e', 'DEMANDED', null, '29292900');
INSERT INTO invoice VALUES (23, 29, 29001, 'AD', 'AD292929/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-05-11T12:01:59+00:00', 2);

-- refunds e2e
INSERT INTO finance_client VALUES (30001, 30, 'refundse2e', 'DEMANDED', null, '30303000');
INSERT INTO invoice VALUES (24, 30, 30001, 'AD', 'AD303030/24', '2024-04-01', '2025-03-31', 10000, null, '2025-03-31', 10, '2024-04-01', null, null, null, '2024-05-11T12:01:59+00:00', 2);
INSERT INTO ledger VALUES (18, 'refund-e2e-credit', '2024-06-11T08:36:40+00:00', '', 15000, '', 'ONLINE CARD PAYMENT', 'CONFIRMED', 30001, null, null, '2024-06-11', '2024-06-12', 1, '', '', 1, '2024-07-05', 2);
INSERT INTO ledger_allocation VALUES (22, 18, 24, '2024-06-11T08:36:40+00:00', 10000, 'ALLOCATED', null, 'refund-e2e-credit', '2024-06-11', 2);
INSERT INTO ledger_allocation VALUES (23, 18, null, '2024-06-11T08:36:40+00:00', -5000, 'UNAPPLIED', null, 'refund-e2e-credit', '2024-06-11', 2);

-- TEST CLIENT DATA: Add data for default client here

-- UPDATE SEQUENCES
SELECT setval('finance_client_id_seq', (SELECT MAX(id) FROM finance_client));
SELECT setval('fee_reduction_id_seq', (SELECT MAX(id) FROM fee_reduction));
SELECT setval('invoice_id_seq', (SELECT MAX(id) FROM invoice));
SELECT setval('invoice_adjustment_id_seq', (SELECT MAX(id) FROM invoice_adjustment));
SELECT setval('ledger_id_seq', (SELECT MAX(id) FROM ledger));
SELECT setval('ledger_allocation_id_seq', (SELECT MAX(id) FROM ledger_allocation));
SELECT setval('invoice_fee_range_id_seq', (SELECT MAX(id) FROM invoice_fee_range));
SELECT setval('refund_id_seq', (SELECT MAX(id) FROM refund));
SELECT setval('pending_collection_id_seq', (SELECT MAX(id) FROM pending_collection));
