package server

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAddInvoiceAdjustmentForm(t *testing.T) {
	permittedAdjustments := []shared.AdjustmentType{shared.AdjustmentTypeDebitMemo, shared.AdjustmentTypeCreditMemo}
	client := mockApiClient{adjustmentTypes: permittedAdjustments}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")
	r.SetPathValue("invoiceId", "9")

	appVars := AppVars{Path: "/path/"}

	sut := AddInvoiceAdjustmentFormHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := AddInvoiceAdjustmentForm{
		&permittedAdjustments,
		"1",
		"9",
		appVars,
	}
	assert.Equal(t, expected, ro.data)
}
