package service

func (suite *IntegrationSuite) Test_AddCollectedPayments() {
	//ctx := suite.ctx
	//seeder := suite.cm.Seeder(ctx, suite.T())
	//
	//seeder.SeedData(
	//	"INSERT INTO finance_client VALUES (1, 1, 'invoice-1', 'DEMANDED', NULL, '1234');",
	//	"INSERT INTO finance_client VALUES (2, 2, 'invoice-2', 'DEMANDED', NULL, '12345');",
	//	"INSERT INTO finance_client VALUES (3, 3, 'invoice-3', 'DEMANDED', NULL, '123456');",
	//	"INSERT INTO finance_client VALUES (4, 4, 'invoice-4', 'DEMANDED', NULL, '1234567');",
	//	"INSERT INTO invoice VALUES (1, 1, 1, 'AD', 'AD11223/19', '2023-04-01', '2025-03-31', 15000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	//	"INSERT INTO invoice VALUES (2, 2, 2, 'AD', 'AD11224/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	//	"INSERT INTO invoice VALUES (3, 3, 3, 'AD', 'AD11225/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	//	"INSERT INTO invoice VALUES (4, 3, 3, 'AD', 'AD11226/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	//	"INSERT INTO invoice VALUES (5, 4, 4, 'AD', 'AD11227/19', '2023-04-01', '2025-03-31', 10000, NULL, '2024-03-31', 11, '2024-03-31', NULL, NULL, NULL, '2024-03-31 00:00:00', '99');",
	//	"INSERT INTO ledger VALUES (1, 'ref', '2024-01-01 15:30:27', '', 10000, 'payment', 'MOTO CARD PAYMENT', 'CONFIRMED', 4, NULL, NULL, NULL, '2024-01-01', NULL, NULL, NULL, NULL, '2020-05-05', 1);",
	//	"INSERT INTO ledger_allocation VALUES (1, 1, 5, '2024-01-01 15:30:27', 10000, 'ALLOCATED', NULL, '', '2024-01-01', NULL);",
	//	"ALTER SEQUENCE ledger_id_seq RESTART WITH 2;",
	//	"ALTER SEQUENCE ledger_allocation_id_seq RESTART WITH 2;",
	//)
	//
	//tests := []struct {
	//	name                      string
	//	records                   [][]string
	//	paymentType               shared.ReportUploadType
	//	bankDate                  shared.Date
	//	pisNumber                 int
	//	expectedClientId          int
	//	expectedLedgerAllocations []createdLedgerAllocation
	//	expectedFailedLines       map[int]string
	//	expectedDispatch          any
	//	want                      error
	//}{
	//	{
	//		name: "Underpayment",
	//		records: [][]string{
	//			{"9800000000000000000", "1234", "100", "D", "01/01/2024"},
	//		},
	//		paymentType:      shared.ReportTypeUploadDirectDebitsCollections,
	//		bankDate:         shared.NewDate("2024-01-17"),
	//		expectedClientId: 1,
	//		expectedLedgerAllocations: []createdLedgerAllocation{
	//			{
	//				10000,
	//				"DIRECT DEBIT PAYMENT",
	//				"CONFIRMED",
	//				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	//				10000,
	//				"ALLOCATED",
	//				1,
	//				0,
	//			},
	//		},
	//		expectedFailedLines: map[int]string{},
	//	},
	//	{
	//		name: "Overpayment",
	//		records: [][]string{
	//			{"Case number (confirmed on Sirius)", "Cheque number", "Cheque Value (£)", "Comments", "Date in Bank"},
	//			{"12345", "54321", "250.10", "", "01/01/2024"},
	//		},
	//		paymentType:      shared.ReportTypeUploadPaymentsSupervisionCheque,
	//		bankDate:         shared.NewDate("2024-01-17"),
	//		pisNumber:        150,
	//		expectedClientId: 2,
	//		expectedLedgerAllocations: []createdLedgerAllocation{
	//			{
	//				25010,
	//				"SUPERVISION CHEQUE PAYMENT",
	//				"CONFIRMED",
	//				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	//				10000,
	//				"ALLOCATED",
	//				2,
	//				150,
	//			},
	//			{
	//				25010,
	//				"SUPERVISION CHEQUE PAYMENT",
	//				"CONFIRMED",
	//				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	//				-15010,
	//				"UNAPPLIED",
	//				0,
	//				150,
	//			},
	//		},
	//		expectedFailedLines: map[int]string{},
	//		expectedDispatch:    event.CreditOnAccount{ClientID: 2, CreditRemaining: 15010},
	//	},
	//	{
	//		name: "Underpayment with multiple invoices",
	//		records: [][]string{
	//			{"Ordercode", "Date", "Amount"},
	//			{"123456", "01/01/2024", "50"},
	//		},
	//		paymentType:      shared.ReportTypeUploadPaymentsMOTOCard,
	//		bankDate:         shared.NewDate("2024-01-17"),
	//		expectedClientId: 3,
	//		expectedLedgerAllocations: []createdLedgerAllocation{
	//			{
	//				5000,
	//				"MOTO CARD PAYMENT",
	//				"CONFIRMED",
	//				time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	//				5000,
	//				"ALLOCATED",
	//				3,
	//				0,
	//			},
	//		},
	//		expectedFailedLines: map[int]string{},
	//	},
	//	{
	//		name: "failure cases",
	//		records: [][]string{
	//			{"Ordercode", "Date", "Amount"},
	//			{"1234567890", "01/01/2024", "50"}, // client not found
	//			{"1234567", "01/01/2024", "100"},   // duplicate
	//		},
	//		paymentType:      shared.ReportTypeUploadPaymentsMOTOCard,
	//		bankDate:         shared.NewDate("2024-01-01"),
	//		expectedClientId: 3,
	//		expectedFailedLines: map[int]string{
	//			1: "CLIENT_NOT_FOUND",
	//			2: "DUPLICATE_PAYMENT",
	//		},
	//	},
	//}
	//for _, tt := range tests {
	//	suite.T().Run(tt.name, func(t *testing.T) {
	//		dispatch := &mockDispatch{}
	//		s := NewService(seeder.Conn, dispatch, nil, nil, nil)
	//
	//		var currentLedgerId int
	//		_ = seeder.QueryRow(suite.ctx, `SELECT MAX(id) FROM ledger`).Scan(&currentLedgerId)
	//
	//		var failedLines map[int]string
	//		failedLines, err := s.ProcessPayments(suite.ctx, tt.records, tt.paymentType, tt.bankDate, tt.pisNumber)
	//		assert.Equal(t, tt.want, err)
	//		assert.Equal(t, tt.expectedFailedLines, failedLines)
	//
	//		var createdLedgerAllocations []createdLedgerAllocation
	//
	//		rows, _ := seeder.Query(suite.ctx,
	//			`SELECT l.amount, l.type, l.status, l.datetime, la.amount, la.status, COALESCE(l.pis_number, 0), COALESCE(la.invoice_id, 0)
	//					FROM ledger l
	//					JOIN ledger_allocation la ON l.id = la.ledger_id
	//				WHERE l.finance_client_id = $1 AND l.id > $2`, tt.expectedClientId, currentLedgerId)
	//
	//		for rows.Next() {
	//			var r createdLedgerAllocation
	//			_ = rows.Scan(&r.ledgerAmount, &r.ledgerType, &r.ledgerStatus, &r.datetime, &r.allocationAmount, &r.allocationStatus, &r.pisNumber, &r.invoiceId)
	//			createdLedgerAllocations = append(createdLedgerAllocations, r)
	//		}
	//
	//		assert.Equal(t, tt.expectedLedgerAllocations, createdLedgerAllocations)
	//		assert.Equal(t, tt.expectedDispatch, dispatch.event)
	//	})
	//}
}
