package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestInvoice(t *testing.T) {
	data := shared.Invoices{
		shared.Invoice{
			Id:                 3,
			Ref:                "N2000001/20",
			Status:             "Unpaid",
			Amount:             232,
			RaisedDate:         shared.NewDate("01/04/2222"),
			Received:           22,
			OutstandingBalance: 210,
			Ledgers: []shared.Ledger{
				{
					Amount:          12300,
					ReceivedDate:    shared.NewDate("01/05/2222"),
					TransactionType: "Online card payment",
					Status:          "APPLIED",
				},
			},
			SupervisionLevels: []shared.SupervisionLevel{
				{
					Level:  "GENERAL",
					Amount: 32000,
					From:   shared.NewDate("01/04/2019"),
					To:     shared.NewDate("31/03/2020"),
				},
			},
		},
	}

	client := mockApiClient{Invoices: data}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := InvoicesHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	out := Invoices{
		{
			Id:                 3,
			Ref:                "N2000001/20",
			Status:             "Unpaid",
			Amount:             232,
			RaisedDate:         "01/04/2222",
			Received:           22,
			OutstandingBalance: 210,
			Ledgers: []LedgerAllocation{
				{
					Amount:          12300,
					ReceivedDate:    shared.NewDate("01/05/2222"),
					TransactionType: "Online Card Payment",
					Status:          "Applied",
				},
			},
			SupervisionLevels: []SupervisionLevel{
				{
					Level:  "General",
					Amount: 32000,
					From:   shared.NewDate("01/04/2019"),
					To:     shared.NewDate("31/03/2020"),
				},
			},
			ClientId: 1,
		},
	}

	expected := &InvoicesVars{
		Invoices: out,
		ClientId: "1",
		AppVars:  appVars,
	}

	assert.Equal(t, expected, ro.data)
}

func TestInvoiceErrors(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := InvoicesHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Equal(t, "this has failed", err.Error())
	assert.False(t, ro.executed)
}

func Test_translate(t *testing.T) {
	tests := []struct {
		name       string
		ledgerType string
		status     string
		amount     int
		want       string
	}{
		{
			name:       "returns a value for something that does not match",
			ledgerType: "THiS wILL not MatcH Any Of thEm",
			want:       "This Will Not Match Any Of Them",
		},
		{
			name:       "returns nothing for no value given in",
			ledgerType: "",
			want:       "",
		},
		{
			name:       "returns a correct value for CREDIT WRITE OFF",
			ledgerType: shared.AdjustmentTypeWriteOff.Key(),
			want:       "Write Off",
		},
		{
			name:       "returns a correct value for CREDIT MEMO",
			ledgerType: shared.AdjustmentTypeCreditMemo.Key(),
			want:       "Manual Credit",
		},
		{
			name:       "returns a correct value for DEBIT MEMO",
			ledgerType: shared.AdjustmentTypeDebitMemo.Key(),
			want:       "Manual Debit",
		},
		{
			name:       "translates payment types",
			ledgerType: shared.TransactionTypeSupervisionBACSPayment.Key(),
			amount:     32000,
			want:       "BACS payment (Supervision account)",
		},
		{
			name:       "translates reversals",
			ledgerType: shared.TransactionTypeSupervisionBACSPayment.Key(),
			amount:     -32000,
			want:       "BACS payment (Supervision account) reversal",
		},
		{
			name:       "returns a correct value for UNAPPLIED",
			ledgerType: shared.FeeReductionTypeHardship.Key(),
			amount:     -1000,
			status:     "UNAPPLIED",
			want:       "Unapplied Payment",
		},
		{
			name:       "returns a correct value for REAPPLIED",
			ledgerType: shared.FeeReductionTypeRemission.Key(),
			amount:     -1000,
			status:     "REAPPLIED",
			want:       "Reapplied Payment",
		},
		{
			name:       "returns Refund Reversal",
			ledgerType: shared.TransactionTypeRefund.Key(),
			amount:     -32000,
			status:     "UNAPPLIED",
			want:       "Refund reversal",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, translate(tt.ledgerType, tt.status, tt.amount), "translate(%v)", tt.ledgerType)
		})
	}
}

func TestInvoicesHandler_transform(t *testing.T) {
	type args struct {
		in       shared.Invoices
		clientId int
	}
	tests := []struct {
		name string
		args args
		want Invoices
	}{
		{
			name: "Returns invoices in the correct order of most recent raised date first",
			args: args{
				in: shared.Invoices{
					shared.Invoice{
						Id:                 3,
						Ref:                "N2000001/33",
						Status:             "Unpaid",
						Amount:             232,
						RaisedDate:         shared.NewDate("01/04/3333"),
						Received:           22,
						OutstandingBalance: 210,
						Ledgers: []shared.Ledger{
							{
								Amount:          12300,
								ReceivedDate:    shared.NewDate("01/05/2222"),
								TransactionType: "Online card payment",
								Status:          "APPLIED",
							},
						},
						SupervisionLevels: []shared.SupervisionLevel{
							{
								Level:  "GENERAL",
								Amount: 32000,
								From:   shared.NewDate("01/04/2019"),
								To:     shared.NewDate("31/03/2020"),
							},
						},
					},
					shared.Invoice{
						Id:                 2,
						Ref:                "N2000001/22",
						Status:             "Unpaid",
						Amount:             232,
						RaisedDate:         shared.NewDate("01/04/2222"),
						Received:           22,
						OutstandingBalance: 210,
						Ledgers: []shared.Ledger{
							{
								Amount:          12300,
								ReceivedDate:    shared.NewDate("01/05/2222"),
								TransactionType: "Online card payment",
								Status:          "APPLIED",
							},
						},
						SupervisionLevels: []shared.SupervisionLevel{
							{
								Level:  "GENERAL",
								Amount: 32000,
								From:   shared.NewDate("01/04/2019"),
								To:     shared.NewDate("31/03/2020"),
							},
						},
					},
					shared.Invoice{
						Id:                 1,
						Ref:                "N2000001/11",
						Status:             "Unpaid",
						Amount:             232,
						RaisedDate:         shared.NewDate("01/04/1111"),
						Received:           22,
						OutstandingBalance: 210,
						Ledgers: []shared.Ledger{
							{
								Amount:          12300,
								ReceivedDate:    shared.NewDate("01/05/2222"),
								TransactionType: "Online card payment",
								Status:          "APPLIED",
							},
						},
						SupervisionLevels: []shared.SupervisionLevel{
							{
								Level:  "GENERAL",
								Amount: 32000,
								From:   shared.NewDate("01/04/2019"),
								To:     shared.NewDate("31/03/2020"),
							},
						},
					},
				},
				clientId: 1,
			},
			want: Invoices{
				{
					Id:                 3,
					Ref:                "N2000001/33",
					Status:             "Unpaid",
					Amount:             232,
					RaisedDate:         "01/04/3333",
					Received:           22,
					OutstandingBalance: 210,
					Ledgers: []LedgerAllocation{
						{
							Amount:          12300,
							ReceivedDate:    shared.NewDate("01/05/2222"),
							TransactionType: "Online Card Payment",
							Status:          "Applied",
						},
					},
					SupervisionLevels: []SupervisionLevel{
						{
							Level:  "General",
							Amount: 32000,
							From:   shared.NewDate("01/04/2019"),
							To:     shared.NewDate("31/03/2020"),
						},
					},
					ClientId: 1,
				},
				{
					Id:                 2,
					Ref:                "N2000001/22",
					Status:             "Unpaid",
					Amount:             232,
					RaisedDate:         "01/04/2222",
					Received:           22,
					OutstandingBalance: 210,
					Ledgers: []LedgerAllocation{
						{
							Amount:          12300,
							ReceivedDate:    shared.NewDate("01/05/2222"),
							TransactionType: "Online Card Payment",
							Status:          "Applied",
						},
					},
					SupervisionLevels: []SupervisionLevel{
						{
							Level:  "General",
							Amount: 32000,
							From:   shared.NewDate("01/04/2019"),
							To:     shared.NewDate("31/03/2020"),
						},
					},
					ClientId: 1,
				},
				{
					Id:                 1,
					Ref:                "N2000001/11",
					Status:             "Unpaid",
					Amount:             232,
					RaisedDate:         "01/04/1111",
					Received:           22,
					OutstandingBalance: 210,
					Ledgers: []LedgerAllocation{
						{
							Amount:          12300,
							ReceivedDate:    shared.NewDate("01/05/2222"),
							TransactionType: "Online Card Payment",
							Status:          "Applied",
						},
					},
					SupervisionLevels: []SupervisionLevel{
						{
							Level:  "General",
							Amount: 32000,
							From:   shared.NewDate("01/04/2019"),
							To:     shared.NewDate("31/03/2020"),
						},
					},
					ClientId: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &InvoicesHandler{
				route{
					client:  nil,
					tmpl:    nil,
					partial: "",
				},
			}
			assert.Equalf(t, tt.want, h.transform(tt.args.in, tt.args.clientId), "transform(%v, %v)", tt.args.in, tt.args.clientId)
		})
	}
}
