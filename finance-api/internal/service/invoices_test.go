package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_GetInvoices() {
	conn := suite.testDB.GetConn()

	conn.SeedData(
		"INSERT INTO finance_client VALUES (7, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO finance_client VALUES (3, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (2, 7, 'REMISSION', NULL, '2019-04-01'::DATE, '2020-03-31'::DATE, 'notes', FALSE, '2019-05-01'::DATE);",
		"INSERT INTO invoice VALUES (1, 1, 7, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', 99);",
		"INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T08:36:40+00:00', '', 12300, '', 'CARD PAYMENT', 'APPROVED', 7, 1, 2, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 65);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T08:36:40+00:00', 12300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_fee_range VALUES (1, 1, 'GENERAL', '2022-04-01', '2023-03-31', 32000);",
	)

	Store := store.New(conn)
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	tests := []struct {
		name    string
		id      int
		want    *shared.Invoices
		wantErr bool
	}{
		{
			name: "returns invoices when clientId matches clientId in invoice table",
			id:   1,
			want: &shared.Invoices{
				shared.Invoice{
					Id:                 1,
					Ref:                "S203531/19",
					Status:             "Unpaid",
					Amount:             32000,
					RaisedDate:         shared.Date{Time: date},
					Received:           12300,
					OutstandingBalance: 19700,
					Ledgers: []shared.Ledger{
						{
							Amount:          12300,
							ReceivedDate:    shared.NewDate("04/12/2022"),
							TransactionType: "CARD PAYMENT",
							Status:          "ALLOCATED",
						},
					},
					SupervisionLevels: []shared.SupervisionLevel{
						{
							Level:  "GENERAL",
							Amount: 32000,
							From:   shared.NewDate("01/04/2022"),
							To:     shared.NewDate("31/03/2023"),
						},
					},
				},
			},
		},
		{
			name: "returns an empty array when no match is found",
			id:   2,
			want: &shared.Invoices{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetInvoices(tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(*tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvoices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_invoiceBuilder_addLedgerAllocations(t *testing.T) {
	tests := []struct {
		name    string
		ilas    []store.GetLedgerAllocationsRow
		status  string
		balance int
	}{
		{
			name:    "Unpaid - no ledgers",
			ilas:    []store.GetLedgerAllocationsRow{},
			status:  "Unpaid",
			balance: 32000,
		},
		{
			name: "Unpaid",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "APPROVED",
				},
			},
			status:  "Unpaid",
			balance: 10000,
		},
		{
			name: "Paid",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "APPROVED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        3,
					Amount:    10000,
					Type:      "CARD PAYMENT",
					Status:    "APPROVED",
				},
			},
			status:  "Paid",
			balance: 0,
		},
		{
			name: "Overpaid",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "APPROVED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        3,
					Amount:    20000,
					Type:      "CARD PAYMENT",
					Status:    "APPROVED",
				},
			},
			status:  "Overpaid",
			balance: -10000,
		},
		{
			name: "Write-off pending",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    22000,
					Type:      "CREDIT WRITE OFF",
					Status:    "PENDING", // ignored for balance but not for status
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        3,
					Amount:    10000,
					Type:      "REMISSION",
					Status:    "APPROVED",
				},
			},
			status:  "Unpaid - Write-off pending",
			balance: 22000,
		},
		{
			name: "Write-off",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    22000,
					Type:      "CREDIT WRITE OFF",
					Status:    "APPROVED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        3,
					Amount:    10000,
					Type:      "REMISSION",
					Status:    "APPROVED",
				},
			},
			status:  "Closed - Write-off",
			balance: 0,
		},
		{
			name: "Closed",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "APPROVED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        3,
					Amount:    10000,
					Type:      "CREDIT MEMO",
					Status:    "APPROVED",
				},
			},
			status:  "Closed",
			balance: 0,
		},
		{
			name: "Remission",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    10000,
					Type:      "REMISSION",
					Status:    "APPROVED",
				},
			},
			status:  "Unpaid - Remission",
			balance: 22000,
		},
		{
			name: "Hardship",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    10000,
					Type:      "HARDSHIP",
					Status:    "APPROVED",
				},
			},
			status:  "Unpaid - Hardship",
			balance: 22000,
		},
		{
			name: "Pending reduction not added as context",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    10000,
					Type:      "HARDSHIP",
					Status:    "PENDING",
				},
			},
			status:  "Unpaid",
			balance: 32000,
		},
		{
			name: "Exemption",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					ID:        2,
					Amount:    10000,
					Type:      "EXEMPTION",
					Status:    "APPROVED",
				},
			},
			status:  "Unpaid - Exemption",
			balance: 22000,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := newInvoiceBuilder([]store.GetInvoicesRow{
				{
					ID:     1,
					Amount: 32000,
				},
			})
			ib.addLedgerAllocations(tt.ilas)
			assert.Equal(t, tt.status, ib.invoices[1].invoice.Status)
			assert.Equal(t, tt.balance, ib.invoices[1].invoice.OutstandingBalance)
		})
	}
}
