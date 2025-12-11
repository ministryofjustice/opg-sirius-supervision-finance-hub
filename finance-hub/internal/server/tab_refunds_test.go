package server

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
)

func TestRefunds(t *testing.T) {
	fulfilledDate := shared.NewDate("02/05/2222")

	data := shared.Refunds{
		CreditBalance: 50,
		Refunds: []shared.Refund{
			{
				ID:            3,
				RaisedDate:    shared.NewDate("01/04/2222"),
				FulfilledDate: shared.NewNillable(&fulfilledDate),
				Amount:        232,
				Status:        shared.RefundStatusFulfilled,
				Notes:         "Some notes here",
				CreatedBy:     99,
				BankDetails: shared.NewNillable(
					&shared.BankDetails{
						Name:     "Billy Banker",
						Account:  "12345678",
						SortCode: "10-20-30",
					},
				),
			},
		},
	}

	client := mockApiClient{refunds: data}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := RefundsHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	out := Refunds{
		{
			ID:            "3",
			DateRaised:    shared.NewDate("01/04/2222"),
			DateFulfilled: &fulfilledDate,
			Amount:        232,
			Status:        "Fulfilled",
			Notes:         "Some notes here",
			CreatedBy:     99,
			BankDetails: &BankDetails{
				Name:     "Billy Banker",
				Account:  "12345678",
				SortCode: "10-20-30",
			},
		},
	}

	expected := &RefundsTab{
		Refunds:       out,
		ShowAddRefund: true,
		ClientId:      "1",
		AppVars:       appVars,
	}

	assert.Equal(t, expected, ro.data)
}

func TestRefunds_Errors(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := RefundsHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Equal(t, "this has failed", err.Error())
	assert.False(t, ro.executed)
}

func TestRefundsTransformStatus(t *testing.T) {
	sut := RefundsHandler{}

	tests := []struct {
		name string
		in   string
		want string
	}{
		{
			"Pending",
			"PENDING",
			"Pending",
		},
		{
			"Rejected",
			"REJECTED",
			"Rejected",
		},
		{
			"Approved",
			"APPROVED",
			"Approved",
		},
		{
			"Processing",
			"PROCESSING",
			"Processing",
		},
		{
			"Cancelled",
			"CANCELLED",
			"Cancelled",
		},
		{
			"Fulfilled",
			"FULFILLED",
			"Fulfilled",
		},
		{
			"Unknown",
			"CONFIRMED",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sut.transformStatus(tt.in), "transformStatus(%v)", tt.in)
		})
	}
}

func TestShouldShowAddRefund(t *testing.T) {
	sut := RefundsHandler{}

	date := shared.NewDate("02/05/2222")
	tests := []struct {
		name string
		in   shared.Refunds
		want bool
	}{
		{
			"contains pending",
			shared.Refunds{
				CreditBalance: 50,
				Refunds: []shared.Refund{
					{
						ID:            2,
						RaisedDate:    shared.NewDate("01/04/2222"),
						FulfilledDate: shared.NewNillable(&date),
						Amount:        232,
						Status:        shared.RefundStatusFulfilled,
						Notes:         "Some notes here",
						CreatedBy:     99,
					},
					{
						ID:         3,
						RaisedDate: shared.NewDate("01/04/2222"),
						Amount:     232,
						Status:     shared.RefundStatusPending,
						Notes:      "Some notes here",
						CreatedBy:  99,
						BankDetails: shared.NewNillable(
							&shared.BankDetails{
								Name:     "Billy Banker",
								Account:  "12345678",
								SortCode: "10-20-30",
							},
						),
					},
				},
			},
			false,
		},
		{
			"approved",
			shared.Refunds{
				CreditBalance: 50,
				Refunds: []shared.Refund{
					{
						ID:         3,
						RaisedDate: shared.NewDate("01/04/2222"),
						Amount:     232,
						Status:     shared.RefundStatusApproved,
						Notes:      "Some notes here",
						CreatedBy:  99,
					},
				},
			},
			false,
		},
		{
			"processing",
			shared.Refunds{
				CreditBalance: 50,
				Refunds: []shared.Refund{
					{
						ID:         3,
						RaisedDate: shared.NewDate("01/04/2222"),
						Amount:     232,
						Status:     shared.RefundStatusProcessing,
						Notes:      "Some notes here",
						CreatedBy:  99,
					},
				},
			},
			false,
		},
		{
			"credit balance",
			shared.Refunds{
				CreditBalance: 0,
				Refunds: []shared.Refund{
					{
						ID:            3,
						RaisedDate:    shared.NewDate("01/04/2222"),
						FulfilledDate: shared.NewNillable(&date),
						Amount:        232,
						Status:        shared.RefundStatusFulfilled,
						Notes:         "Some notes here",
						CreatedBy:     99,
					},
				},
			},
			false,
		},
		{
			"should show add refund",
			shared.Refunds{
				CreditBalance: 50,
				Refunds: []shared.Refund{
					{
						ID:            3,
						RaisedDate:    shared.NewDate("01/04/2222"),
						FulfilledDate: shared.NewNillable(&date),
						Amount:        232,
						Status:        shared.RefundStatusFulfilled,
						Notes:         "Some notes here",
						CreatedBy:     99,
						BankDetails: shared.NewNillable(
							&shared.BankDetails{
								Name:     "Billy Banker",
								Account:  "12345678",
								SortCode: "10-20-30",
							},
						),
					},
					{
						ID:         3,
						RaisedDate: shared.NewDate("01/04/2222"),
						Amount:     232,
						Status:     shared.RefundStatusRejected,
						Notes:      "Some notes here",
						CreatedBy:  99,
					},
					{
						ID:         3,
						RaisedDate: shared.NewDate("01/04/2222"),
						Amount:     232,
						Status:     shared.RefundStatusCancelled,
						Notes:      "Some notes here",
						CreatedBy:  99,
					},
				},
			},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sut.shouldShowAddRefund(tt.in), "shouldShowAddRefund(%v)", tt.in)
		})
	}
}
