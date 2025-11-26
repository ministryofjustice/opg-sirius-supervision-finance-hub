package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

type createdReversalAllocation struct {
	ledgerAmount     int
	ledgerType       string
	ledgerStatus     string
	receivedDate     time.Time
	bankDate         time.Time
	allocationAmount int
	allocationStatus string
	invoiceId        pgtype.Int4
	financeClientId  int
	notes            string
	pisNumber        int
}

func (suite *IntegrationSuite) Test_processReversals() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		// test 1
		"INSERT INTO finance_client VALUES (1, 1, 'test 1', 'DEMANDED', NULL, '1111');",
		"INSERT INTO finance_client VALUES (2, 2, 'test 1 - replacement', 'DEMANDED', NULL, '2222');",
		"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'test 1 paid', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (1, 'client-1-reverse-me', '2024-01-02 15:32:10', '', 1000, 'payment 1', 'MOTO CARD PAYMENT', 'CONFIRMED', 1, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2024-01-02 15:32:10', 1000, 'ALLOCATED', NULL, '', '2024-01-01', NULL);",
		"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'test 1 replacement unpaid', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",

		//test 2
		"INSERT INTO finance_client VALUES (3, 3, 'test 2', 'DEMANDED', NULL, '3333');",
		"INSERT INTO finance_client VALUES (4, 4, 'test 2 - replacement', 'DEMANDED', NULL, '4444');",
		"INSERT INTO invoice VALUES (3, 3, 3, 'AD', 'test 2 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'test 2 partially paid with payment', '2023-05-01', '2025-04-01', 10000, NULL, '2024-04-01', NULL, '2024-04-01', NULL, NULL, NULL, '2024-04-01 00:00:00', '99');",
		"INSERT INTO invoice VALUES (5, 3, 3, 'AD', 'test 2 unpaid with payment', '2023-06-01', '2025-05-01', 10000, NULL, '2025-05-01', NULL, '2025-05-01', NULL, NULL, NULL, '2025-05-01 00:00:00', '99');",
		"INSERT INTO ledger VALUES (2, 'test 2', '2025-01-02 15:32:10', '', 15000, 'payment 2', 'ONLINE CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (2, 2, 3, '2025-01-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (3, 2, 4, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		// second payment received after the payment being reversed
		"INSERT INTO ledger VALUES (3, 'test 2 - second payment', '2025-01-03 15:32:10', '', 10000, 'payment 2 - not reversed', 'MOTO CARD PAYMENT', 'CONFIRMED', 3, NULL, NULL, NULL, '2025-01-03', NULL, NULL, NULL, NULL, '2025-01-03', 1);",
		"INSERT INTO ledger_allocation VALUES (4, 3, 4, '2025-01-03 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-03', NULL);",
		"INSERT INTO ledger_allocation VALUES (5, 3, 5, '2025-01-03 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-03', NULL);",

		// test 3
		"INSERT INTO finance_client VALUES (5, 5, 'test 3', 'DEMANDED', NULL, '5555');",
		"INSERT INTO finance_client VALUES (6, 6, 'test 3 - replacement', 'DEMANDED', NULL, '6666');",
		"INSERT INTO invoice VALUES (6, 5, 5, 'AD', 'test 3 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (4, 'test 3', '2025-01-02 15:32:10', '', 15000, 'payment 3', 'ONLINE CARD PAYMENT', 'CONFIRMED', 5, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (6, 4, 6, '2025-01-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (7, 4, NULL, '2025-01-02 15:32:10', -5000, 'UNAPPLIED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO invoice VALUES (7, 6, 6, 'AD', 'test 3 replacement', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",

		// bounced cheque
		"INSERT INTO finance_client VALUES (7, 7, 'bounced cheque', 'DEMANDED', NULL, '7777');",
		"INSERT INTO invoice VALUES (8, 7, 7, 'AD', 'bounced cheque paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (5, 'bounced cheque', '2025-01-02 15:32:10', '', 10000, 'payment 4', 'SUPERVISION CHEQUE PAYMENT', 'CONFIRMED', 7, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1, 123);",
		"INSERT INTO ledger_allocation VALUES (8, 5, 8, '2025-01-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// duplicate payment
		"INSERT INTO finance_client VALUES (8, 8, 'test 4', 'DEMANDED', NULL, '8888');",
		"INSERT INTO invoice VALUES (9, 8, 8, 'AD', 'test 4 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (6, 'test 4', '2025-01-02 15:32:10', '', 5000, 'payment 4', 'ONLINE CARD PAYMENT', 'CONFIRMED', 8, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (9, 6, 9, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// invalid duplicate reversal, same reversal in previous upload
		"INSERT INTO finance_client VALUES (9, 9, 'test 5', 'DEMANDED', NULL, '9999');",
		"INSERT INTO invoice VALUES (10, 9, 9, 'AD', 'test 5 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (7, 'test 5', '2025-01-02 15:32:10', '', 5000, 'payment 5', 'ONLINE CARD PAYMENT', 'CONFIRMED', 9, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (10, 7, 10, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger VALUES (8, 'test 7', '2025-01-02 15:32:10', '', -5000, 'payment 7', 'ONLINE CARD PAYMENT', 'CONFIRMED', 9, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (11, 8, 10, '2025-01-02 15:32:10', -5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// invalid duplicate reversal, same reversal in upload
		"INSERT INTO finance_client VALUES (10, 10, 'test 6', 'DEMANDED', NULL, '1010');",
		"INSERT INTO invoice VALUES (11, 10, 10, 'AD', 'test 6 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (9, 'test 6', '2025-01-02 15:32:10', '', 5000, 'payment 6', 'ONLINE CARD PAYMENT', 'CONFIRMED', 10, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (12, 9, 11, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// valid duplicate reversal for duplicate payment upload
		"INSERT INTO finance_client VALUES (11, 11, 'test 8', 'DEMANDED', NULL, '1011');",
		"INSERT INTO invoice VALUES (12, 11, 11, 'AD', 'test 8 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (10, 'test 8', '2025-01-02 15:32:10', '', 5000, 'payment 8', 'ONLINE CARD PAYMENT', 'CONFIRMED', 11, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (13, 10, 12, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger VALUES (11, 'test 9', '2025-01-02 15:32:10', '', 5000, 'payment 9', 'ONLINE CARD PAYMENT', 'CONFIRMED', 11, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (14, 11, 12, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// failed direct debit collection
		"INSERT INTO finance_client VALUES (12, 12, 'test 12', 'DEMANDED', NULL, '1212');",
		"INSERT INTO invoice VALUES (13, 12, 12, 'AD', 'test 12 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (12, 'test 12', '2025-01-02 15:32:10', '', 5000, 'payment 12', 'DIRECT DEBIT PAYMENT', 'CONFIRMED', 12, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (15, 12, 13, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// failed reversal due to insufficient debt position
		"INSERT INTO finance_client VALUES (13, 13, 'test 13', 'DEMANDED', NULL, '1313');",
		"INSERT INTO invoice VALUES (14, 13, 13, 'AD', 'test 13 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (13, 'test 13', '2025-01-02 15:32:10', '', 10000, 'payment 13', 'ONLINE CARD PAYMENT', 'CONFIRMED', 13, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (16, 13, 14, '2025-01-02 15:32:10', 10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger VALUES (14, 'test 13 - debit', '2025-01-02 15:32:10', '', -10000, 'manually reverses', 'DEBIT MEMO', 'CONFIRMED', 13, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (17, 14, 14, '2025-01-02 15:32:10', -10000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		// client in credit but payment being reversed only covers invoice
		"INSERT INTO finance_client VALUES (14, 14, 'test 14', 'DEMANDED', NULL, '1414');",
		"INSERT INTO invoice VALUES (15, 14, 14, 'AD', 'test 14 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (15, 'test 14.1', '2025-01-02 15:32:10', '', 5000, 'payment 14 being reversed', 'ONLINE CARD PAYMENT', 'CONFIRMED', 14, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (18, 15, 15, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger VALUES (16, 'test 14.2', '2025-01-02 15:32:10', '', 10000, '14 gone into credit', 'CREDIT MEMO', 'CONFIRMED', 14, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (19, 16, 15, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO ledger_allocation VALUES (20, 16, 15, '2025-01-02 15:32:10', -5000, 'UNAPPLIED', NULL, '', '2025-01-02', NULL);",

		// failed direct debit collection
		"INSERT INTO finance_client VALUES (15, 15, 'test 15', 'DEMANDED', NULL, '1515');",
		"INSERT INTO invoice VALUES (16, 15, 15, 'AD', 'test 15 paid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (17, 'test 15', '2025-01-02 15:32:10', '', 5000, 'payment 15', 'REFUND', 'CONFIRMED', 15, NULL, NULL, NULL, '2025-01-01', NULL, NULL, NULL, NULL, '2025-01-02', 1);",
		"INSERT INTO ledger_allocation VALUES (21, 17, 16, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",

		"ALTER SEQUENCE ledger_id_seq RESTART WITH 18;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 22;",
		// misapplied cheque passes through PIS number
		"INSERT INTO finance_client VALUES (16, 16, 'test 16', 'DEMANDED', NULL, '1616');",
		"INSERT INTO invoice VALUES (17, 16, 16, 'AD', 'test 16 unpaid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
		"INSERT INTO ledger VALUES (18, 'test 16', '2025-01-02 15:32:10', '', 5000, 'payment 15 being reversed', 'SUPERVISION CHEQUE PAYMENT', 'CONFIRMED', 16, NULL, NULL, NULL, '2025-01-02', NULL, NULL, NULL, NULL, '2025-01-02', 1, 101);",
		"INSERT INTO ledger_allocation VALUES (22, 18, 17, '2025-01-02 15:32:10', 5000, 'ALLOCATED', NULL, '', '2025-01-02', NULL);",
		"INSERT INTO finance_client VALUES (17, 17, 'test 17', 'DEMANDED', NULL, '1717');",
		"INSERT INTO invoice VALUES (18, 17, 17, 'AD', 'test 17 unpaid', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', NULL, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",

		"ALTER SEQUENCE ledger_id_seq RESTART WITH 19;",
		"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 23;",
	)

	dispatch := &mockDispatch{}
	s := Service{store: store.New(seeder.Conn), dispatch: dispatch, tx: seeder.Conn}
	uploadDate := shared.NewDate("2025-02-03")

	tests := []struct {
		name                string
		records             [][]string
		uploadType          shared.ReportUploadType
		uploadDate          shared.Date
		allocations         []createdReversalAllocation
		expectedFailedLines map[int]string
		want                error
	}{
		{
			name:       "failure cases with eventual success",
			uploadType: shared.ReportTypeUploadMisappliedPayments,
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"MOTO CARD PAYMENT", "0000", "2222", "01/01/2024", "02/01/2024", "10.00", ""},             // current court reference not found
				{"MOTO CARD PAYMENT", "1111", "0000", "01/01/2024", "02/01/2024", "10.00", ""},             // new court reference not found
				{"ONLINE CARD PAYMENT", "1111", "2222", "01/01/2024", "02/01/2024", "10.00", ""},           // incorrect payment type
				{"MOTO CARD PAYMENT", "1111", "2222", "12/01/2024", "02/01/2024", "10.00", ""},             // bank date does not match payment
				{"MOTO CARD PAYMENT", "1111", "2222", "01/01/2024", "12/01/2024", "10.00", ""},             // received date does not match payment
				{"MOTO CARD PAYMENT", "1111", "2222", "01/01/2024", "02/01/2024", "10.01", ""},             // amount does not match payment
				{"MOTO CARD PAYMENT", "1111", "2222", "01/01/2024", "02/01/2024", "10.00", ""},             // successful match
				{"SUPERVISION CHEQUE PAYMENT", "1616", "1717", "02/01/2025", "02/01/2025", "50.00", "101"}, // successful match
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -1000,
					ledgerType:       "MOTO CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2024, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC),
					allocationAmount: -1000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 1, Valid: true},
					financeClientId:  1,
				},
				{
					ledgerAmount:     1000,
					ledgerType:       "MOTO CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2024, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2024, 01, 01, 0, 0, 0, 0, time.UTC),
					allocationAmount: 1000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 2, Valid: true},
					financeClientId:  2,
				},
				{
					ledgerAmount:     -5000,
					ledgerType:       "SUPERVISION CHEQUE PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 17, Valid: true},
					financeClientId:  16,
					pisNumber:        101,
				},
				{
					ledgerAmount:     5000,
					ledgerType:       "SUPERVISION CHEQUE PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: 5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 18, Valid: true},
					financeClientId:  17,
					pisNumber:        101,
				},
			},
			expectedFailedLines: map[int]string{
				1: "NO_MATCHED_PAYMENT",
				2: "REVERSAL_CLIENT_NOT_FOUND",
				3: "NO_MATCHED_PAYMENT",
				4: "NO_MATCHED_PAYMENT",
				5: "NO_MATCHED_PAYMENT",
				6: "NO_MATCHED_PAYMENT",
			},
		},
		{
			name: "misapplied payment - original payment over two invoices applied to client with overpayment",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "3333", "4444", "02/01/2025", "02/01/2025", "150.00", ""},
			},
			uploadType: shared.ReportTypeUploadMisappliedPayments,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 5, Valid: true},
					financeClientId:  3,
				},
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 4, Valid: true},
					financeClientId:  3,
				},
				{
					ledgerAmount:     15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -15000, // unapply so a negative amount is allocated
					allocationStatus: "UNAPPLIED",
					invoiceId:        pgtype.Int4{}, // no invoice on replacement client so unapplied as overpayment
					financeClientId:  4,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "misapplied payment - errored client in credit",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "New (correct) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "5555", "6666", "02/01/2025", "02/01/2025", "150.00", ""},
			},
			uploadType: shared.ReportTypeUploadMisappliedPayments,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: 5000, // positive unapply to reverse the existing credit balance
					allocationStatus: "UNAPPLIED",
					invoiceId:        pgtype.Int4{},
					financeClientId:  5,
				},
				{
					ledgerAmount:     -15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 6, Valid: true},
					financeClientId:  5,
				},
				{
					ledgerAmount:     15000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: 15000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 7, Valid: true},
					financeClientId:  6,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "bounced cheque",
			records: [][]string{
				{"Court reference", "Bank date", "Received date", "Amount", "PIS number"},
				{"7777", "02/01/2025", "02/01/2025", "100.00", "123"},
			},
			uploadType: shared.ReportTypeUploadBouncedCheque,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -10000,
					ledgerType:       "SUPERVISION CHEQUE PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -10000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 8, Valid: true},
					financeClientId:  7,
					pisNumber:        123,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "duplicate payment",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "8888", "02/01/2025", "02/01/2025", "50.00", ""},
			},
			uploadType: shared.ReportTypeUploadDuplicatedPayments,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -5000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 9, Valid: true},
					financeClientId:  8,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "duplicate reversal in new file should throw error",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "9999", "02/01/2025", "02/01/2025", "50.00", ""},
			},
			uploadType:          shared.ReportTypeUploadDuplicatedPayments,
			expectedFailedLines: map[int]string{1: "DUPLICATE_REVERSAL"},
		},
		{
			name: "duplicate reversal in same file should throw error",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "1010", "02/01/2025", "02/01/2025", "50.00", ""},
				{"ONLINE CARD PAYMENT", "1010", "02/01/2025", "02/01/2025", "50.00", ""},
			},
			uploadType: shared.ReportTypeUploadDuplicatedPayments,
			expectedFailedLines: map[int]string{
				2: "DUPLICATE_REVERSAL",
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -5000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 11, Valid: true},
					financeClientId:  10,
				},
			},
		},
		{
			name: "reversal of duplicate payments should be allowed",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "1011", "02/01/2025", "02/01/2025", "50.00", ""},
				{"ONLINE CARD PAYMENT", "1011", "02/01/2025", "02/01/2025", "50.00", ""},
				{"ONLINE CARD PAYMENT", "1011", "02/01/2025", "02/01/2025", "50.00", ""},
			},
			uploadType: shared.ReportTypeUploadDuplicatedPayments,
			expectedFailedLines: map[int]string{
				3: "DUPLICATE_REVERSAL",
			},
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -5000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 12, Valid: true},
					financeClientId:  11,
				},
				{
					ledgerAmount:     -5000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 00, 00, 00, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 12, Valid: true},
					financeClientId:  11,
				},
			},
		},
		{
			name: "failed direct debit collection",
			records: [][]string{
				{"Court reference", "Bank date", "Received date", "Amount"},
				{"1212", "13/11/2025", "02/01/2025", "50.00"}, // different bank date as this should not be matched on
			},
			uploadType: shared.ReportTypeUploadFailedDirectDebitCollections,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -5000,
					ledgerType:       "DIRECT DEBIT PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 11, 13, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 13, Valid: true},
					financeClientId:  12,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "unable to reverse due to insufficient debt position",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "1313", "02/01/2025", "02/01/2025", "100.00", ""},
			},
			uploadType: shared.ReportTypeUploadDuplicatedPayments,
			expectedFailedLines: map[int]string{
				1: "MAXIMUM_DEBT",
			},
		},
		{
			name: "applies reversal to credit first",
			records: [][]string{
				{"Payment type", "Current (errored) court reference", "Bank date", "Received date", "Amount", "PIS number (cheque only)"},
				{"ONLINE CARD PAYMENT", "1414", "02/01/2025", "02/01/2025", "50.00", ""},
			},
			uploadType: shared.ReportTypeUploadDuplicatedPayments,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -5000,
					ledgerType:       "ONLINE CARD PAYMENT",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 01, 02, 0, 0, 0, 0, time.UTC),
					allocationAmount: 5000,
					allocationStatus: "UNAPPLIED",
					financeClientId:  14,
				},
			},
			expectedFailedLines: map[int]string{},
		},
		{
			name: "refund reversal",
			records: [][]string{
				{"Court reference", "Amount", "Bank date"},
				{"1515", "50.00", "01/01/2025"},
			},
			uploadType: shared.ReportTypeUploadReverseFulfilledRefunds,
			uploadDate: uploadDate,
			allocations: []createdReversalAllocation{
				{
					ledgerAmount:     -5000,
					ledgerType:       "REFUND",
					ledgerStatus:     "CONFIRMED",
					receivedDate:     time.Date(2025, 02, 03, 0, 0, 0, 0, time.UTC),
					bankDate:         time.Date(2025, 02, 03, 0, 0, 0, 0, time.UTC),
					allocationAmount: -5000,
					allocationStatus: "ALLOCATED",
					invoiceId:        pgtype.Int4{Int32: 16, Valid: true},
					financeClientId:  15,
				},
			},
			expectedFailedLines: map[int]string{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			var currentLedgerId int
			_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)

			var failedLines map[int]string
			failedLines, err := s.ProcessPaymentReversals(suite.ctx, tt.records, tt.uploadType, tt.uploadDate)

			assert.Equal(t, tt.want, err)
			assert.Equal(t, tt.expectedFailedLines, failedLines)

			var allocations []createdReversalAllocation

			rows, _ := seeder.Query(suite.ctx,
				`SELECT l.amount, l.type, l.status, l.datetime, l.bankdate, la.amount, la.status, l.finance_client_id, la.invoice_id, COALESCE(l.pis_number, 0)
						FROM ledger l
						LEFT JOIN ledger_allocation la ON l.id = la.ledger_id
					WHERE l.id > $1`, currentLedgerId)

			for rows.Next() {
				var r createdReversalAllocation
				err := rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.receivedDate, &r.bankDate, &r.allocationAmount, &r.allocationStatus, &r.financeClientId, &r.invoiceId, &r.pisNumber)
				if err != nil {
					fmt.Println(err.Error())
				}
				allocations = append(allocations, r)
			}

			assert.Equal(t, tt.allocations, allocations)
		})
	}
}
