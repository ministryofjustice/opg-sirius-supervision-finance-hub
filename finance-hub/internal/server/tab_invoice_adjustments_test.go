package server

import (
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPendingInvoiceAdjustments(t *testing.T) {
	data := shared.InvoiceAdjustments{
		shared.InvoiceAdjustment{
			Id:             3,
			InvoiceRef:     "N2000001/20",
			Status:         "PENDING",
			Amount:         232,
			RaisedDate:     shared.NewDate("01/04/2222"),
			AdjustmentType: shared.AdjustmentTypeCreditMemo,
			Notes:          "Some notes",
		},
	}

	client := mockApiClient{invoiceAdjustments: data}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := InvoiceAdjustmentsHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	out := InvoiceAdjustments{
		PendingInvoiceAdjustment{
			Id:               "3",
			Invoice:          "N2000001/20",
			Status:           "Pending",
			AdjustmentAmount: "2.32",
			DateRaised:       shared.NewDate("01/04/2222"),
			AdjustmentType:   "Credit",
			Notes:            "Some notes",
		},
	}

	expected := &InvoiceAdjustmentsTab{
		InvoiceAdjustments: out,
		ClientId:           "1",
		AppVars:            appVars,
	}

	assert.Equal(t, expected, ro.data)
}

func TestPendingInvoiceAdjustments_Errors(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("this has failed")
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	appVars := AppVars{Path: "/path/"}

	sut := InvoiceAdjustmentsHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Equal(t, "this has failed", err.Error())
	assert.False(t, ro.executed)
}

func TestPendingInvoiceAdjustmentsTransformType(t *testing.T) {
	sut := InvoiceAdjustmentsHandler{}

	tests := []struct {
		name string
		in   shared.AdjustmentType
		want string
	}{
		{
			"Credit",
			shared.AdjustmentTypeCreditMemo,
			"Credit",
		},
		{
			"Debit",
			shared.AdjustmentTypeDebitMemo,
			"Debit",
		},
		{
			"Write off",
			shared.AdjustmentTypeWriteOff,
			"Write off",
		},
		{
			"Unknown",
			shared.AdjustmentTypeUnknown,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, sut.transformType(tt.in), "transformType(%v)", tt.in)
		})
	}
}

func TestPendingInvoiceAdjustmentsTransformStatus(t *testing.T) {
	sut := InvoiceAdjustmentsHandler{}

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
