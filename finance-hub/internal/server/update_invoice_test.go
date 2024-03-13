package server

import (
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUpdateInvoice(t *testing.T) {
	client := mockApiClient{}
	ro := &mockRoute{client: client}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("id", "1")
	r.SetPathValue("invoiceId", "9")

	appVars := AppVars{Path: "/path/"}

	sut := UpdateInvoiceHandler{ro}
	err := sut.render(appVars, w, r)

	assert.Nil(t, err)
	assert.True(t, ro.executed)

	expected := UpdateInvoices{
		shared.InvoiceTypes,
		"1",
		"9",
		appVars,
	}
	assert.Equal(t, expected, ro.data)
}
