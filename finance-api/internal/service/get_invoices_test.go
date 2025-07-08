package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-api/internal/store"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
	"time"
)

func (suite *IntegrationSuite) TestService_GetInvoices() {
	ctx := suite.ctx
	seeder := suite.cm.Seeder(ctx, suite.T())

	seeder.SeedData(
		"INSERT INTO finance_client VALUES (7, 1, '1234', 'DEMANDED', NULL);",
		"INSERT INTO finance_client VALUES (3, 2, '1234', 'DEMANDED', NULL);",
		"INSERT INTO fee_reduction VALUES (1, 7, 'REMISSION', NULL, '2019-04-01'::DATE, '2020-03-31'::DATE, 'notes', FALSE, '2019-05-01'::DATE);",
		"INSERT INTO invoice VALUES (1, 1, 7, 'S2', 'S203531/19', '2019-04-01', '2020-03-31', 32000, NULL, '2020-03-20',1, '2020-03-16', 10, NULL, NULL, '2019-06-06', 99);",
		"INSERT INTO ledger VALUES (1, 'random1223', '2022-04-11T00:00:00+00:00', '', 12300, '', 'CREDIT REMISSION', 'CONFIRMED', 7, 1, 1, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger_allocation VALUES (1, 1, 1, '2022-04-11T00:00:00+00:00', 12300, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO ledger_allocation VALUES (2, 1, 1, '2022-04-11T00:00:00+00:00', -2300, 'UNAPPLIED', NULL, 'Notes here', '2022-04-11', NULL);",
		"INSERT INTO invoice_fee_range VALUES (1, 1, 'GENERAL', '2022-04-01', '2023-03-31', 32000);",
		// this ledger and allocation should be ignored as ledger is APPROVED, not CONFIRMED (PFS-206)
		"INSERT INTO ledger VALUES (2, 'ignore', '2022-04-11T08:36:40+00:00', '', 99999, '', 'CREDIT MEMO', 'APPROVED', 7, NULL, NULL, '11/04/2022', '12/04/2022', 1254, '', '', 1, '05/05/2022', 2);",
		"INSERT INTO ledger_allocation VALUES (3, 2, 1, '2022-04-11T08:36:40+00:00', 99999, 'ALLOCATED', NULL, 'Notes here', '2022-04-11', NULL);",
	)

	Store := store.New(seeder)
	dateString := "2020-03-16"
	date, _ := time.Parse("2006-01-02", dateString)
	tests := []struct {
		name    string
		id      int32
		want    shared.Invoices
		wantErr bool
	}{
		{
			name: "returns invoices when clientId matches clientId in invoice table",
			id:   1,
			want: shared.Invoices{
				shared.Invoice{
					Id:                 1,
					Ref:                "S203531/19",
					Status:             "Unpaid - Remission",
					Amount:             32000,
					RaisedDate:         shared.Date{Time: date},
					Received:           10000,
					OutstandingBalance: 22000,
					Ledgers: []shared.Ledger{
						{
							Amount:          -2300,
							ReceivedDate:    shared.NewDate("11/04/2022"),
							TransactionType: "CREDIT REMISSION",
							Status:          "UNAPPLIED",
						},
						{
							Amount:          12300,
							ReceivedDate:    shared.NewDate("11/04/2022"),
							TransactionType: "CREDIT REMISSION",
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
			want: shared.Invoices{},
		},
	}
	for _, tt := range tests {
		suite.T().Run(tt.name, func(t *testing.T) {
			s := &Service{
				store: Store,
			}
			got, err := s.GetInvoices(suite.ctx, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetInvoices() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (err == nil) && len(tt.want) == 0 {
				assert.Empty(t, got)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetInvoices() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_invoiceBuilder_statuses(t *testing.T) {
	tests := []struct {
		name         string
		ilas         []store.GetLedgerAllocationsRow
		status       string
		balance      int
		feeReduction string
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
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "ALLOCATED",
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
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "CARD PAYMENT",
					Status:    "ALLOCATED",
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
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    20000,
					Type:      "CARD PAYMENT",
					Status:    "ALLOCATED",
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
					Amount:    22000,
					Type:      "CREDIT WRITE OFF",
					Status:    "PENDING", // ignored for balance but not for status
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "REMISSION",
					Status:    "ALLOCATED",
				},
			},
			status:       "Unpaid - Write-off pending",
			balance:      22000,
			feeReduction: "REMISSION",
		},
		{
			name: "Write-off",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    22000,
					Type:      "CREDIT WRITE OFF",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "REMISSION",
					Status:    "ALLOCATED",
				},
			},
			status:       "Closed - Write-off",
			balance:      0,
			feeReduction: "REMISSION",
		},
		{
			name: "Write-off reversed",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    -22000,
					Type:      "WRITE OFF REVERSAL",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    22000,
					Type:      "CREDIT WRITE OFF",
					Status:    "ALLOCATED",
				},
			},
			status:  "Closed",
			balance: 0,
		},
		{
			name: "Closed with unapplied credit",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "CREDIT MEMO",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    32000,
					Type:      "EXEMPTION",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "CREDIT MEMO",
					Status:    "UNAPPLIED",
				},
			},
			status:       "Closed - Exemption",
			balance:      0,
			feeReduction: "EXEMPTION",
		},
		{
			name: "Paid - with credit applied",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    22000,
					Type:      "CARD PAYMENT",
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "CREDIT MEMO",
					Status:    "ALLOCATED",
				},
			},
			status:  "Paid",
			balance: 0,
		},
		{
			name: "Remission",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "REMISSION",
					Status:    "ALLOCATED",
				},
			},
			status:       "Unpaid - Remission",
			balance:      22000,
			feeReduction: "REMISSION",
		},
		{
			name: "Hardship",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "HARDSHIP",
					Status:    "ALLOCATED",
				},
			},
			status:       "Unpaid - Hardship",
			balance:      22000,
			feeReduction: "HARDSHIP",
		},
		{
			name: "Exemption",
			ilas: []store.GetLedgerAllocationsRow{
				{
					InvoiceID: pgtype.Int4{Int32: 1, Valid: true},
					Amount:    10000,
					Type:      "EXEMPTION",
					Status:    "ALLOCATED",
				},
			},
			status:       "Unpaid - Exemption",
			balance:      22000,
			feeReduction: "EXEMPTION",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ib := newInvoiceBuilder([]store.GetInvoicesRow{
				{
					ID:               1,
					Amount:           int32(tt.balance),
					FeeReductionType: tt.feeReduction,
				},
			})
			ib.addLedgerAllocations(tt.ilas)
			invoices := ib.Build()
			assert.Equal(t, tt.status, invoices[0].Status)
		})
	}
}

func Test_invoiceBuilder_OrdersByIndex(t *testing.T) {
	ib2 := &invoiceBuilder{
		invoices:            make(map[int]*invoiceMetadata),
		invoicePositionByID: make(map[int32]int),
	}

	ib2.invoices[5] = &invoiceMetadata{
		invoice: &shared.Invoice{
			Id:                 5,
			Ref:                "REF5",
			Amount:             5000,
			RaisedDate:         shared.Date{Time: time.Date(2023, 5, 5, 0, 0, 0, 0, time.UTC)},
			Status:             "Unpaid",
			Received:           0,
			OutstandingBalance: 5000,
		},
	}

	ib2.invoices[2] = &invoiceMetadata{
		invoice: &shared.Invoice{
			Id:                 2,
			Ref:                "REF2",
			Amount:             2000,
			RaisedDate:         shared.Date{Time: time.Date(2023, 2, 2, 0, 0, 0, 0, time.UTC)},
			Status:             "Unpaid",
			Received:           0,
			OutstandingBalance: 2000,
		},
	}

	ib2.invoices[8] = &invoiceMetadata{
		invoice: &shared.Invoice{
			Id:                 8,
			Ref:                "REF8",
			Amount:             8000,
			RaisedDate:         shared.Date{Time: time.Date(2023, 8, 8, 0, 0, 0, 0, time.UTC)},
			Status:             "Unpaid",
			Received:           0,
			OutstandingBalance: 8000,
		},
	}

	result2 := ib2.Build()

	assert.Equal(t, 3, len(result2))
	assert.Equal(t, 2, result2[0].Id)
	assert.Equal(t, "REF2", result2[0].Ref)
	assert.Equal(t, 5, result2[1].Id)
	assert.Equal(t, "REF5", result2[1].Ref)
	assert.Equal(t, 8, result2[2].Id)
	assert.Equal(t, "REF8", result2[2].Ref)
}
