package service

import (
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/opg-sirius-finance-hub/finance-api/internal/store"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_generateLedgerEntries(t *testing.T) {
	tests := []struct {
		name   string
		args   addLedgerVars
		wantL  store.CreateLedgerParams
		wantLA []store.CreateLedgerAllocationParams
	}{
		{
			name: "No existing payments",
			args: addLedgerVars{
				amount:             10000,
				transactionType:    shared.FeeReductionTypeHardship,
				feeReductionId:     1,
				clientId:           2,
				invoiceId:          3,
				outstandingBalance: 10000,
			},
			wantL: store.CreateLedgerParams{
				ClientID:       2,
				Amount:         10000,
				Notes:          pgtype.Text{String: "Credit due to approved hardship", Valid: true},
				Type:           "CREDIT HARDSHIP",
				Status:         "APPROVED",
				FeeReductionID: pgtype.Int4{Int32: 1, Valid: true},
				CreatedBy:      pgtype.Int4{Int32: 1},
			},
			wantLA: []store.CreateLedgerAllocationParams{
				{
					InvoiceID: pgtype.Int4{Int32: 3, Valid: true},
					Amount:    10000,
					Status:    "ALLOCATED",
					Notes:     pgtype.Text{},
				},
			},
		},
		{
			name: "Unapply excess credit",
			args: addLedgerVars{
				amount:             10000,
				transactionType:    shared.AdjustmentTypeCreditMemo,
				clientId:           2,
				invoiceId:          3,
				outstandingBalance: 3000,
			},
			wantL: store.CreateLedgerParams{
				ClientID:  2,
				Amount:    3000,
				Notes:     pgtype.Text{String: "Credit due to approved credit memo", Valid: true},
				Type:      "CREDIT MEMO",
				Status:    "APPROVED",
				CreatedBy: pgtype.Int4{Int32: 1},
			},
			wantLA: []store.CreateLedgerAllocationParams{
				{
					InvoiceID: pgtype.Int4{Int32: 3, Valid: true},
					Amount:    10000,
					Status:    "ALLOCATED",
				},
				{
					InvoiceID: pgtype.Int4{Int32: 3, Valid: true},
					Amount:    -7000,
					Status:    "UNAPPLIED",
					Notes:     pgtype.Text{String: "Unapplied funds as a result of applying credit memo", Valid: true},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ledgers, allocations := generateLedgerEntries(tt.args)
			assert.Equalf(t, tt.wantL, ledgers, "generateLedgerEntries(%v)", tt.args)
			assert.Equalf(t, tt.wantLA, allocations, "generateLedgerEntries(%v)", tt.args)
		})
	}
}
