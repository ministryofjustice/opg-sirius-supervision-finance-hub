package server

import (
	"github.com/opg-sirius-finance-hub/finance-hub/internal/api"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
)

type mockTemplate struct {
	executed         bool
	executedTemplate bool
	lastVars         interface{}
	lastW            io.Writer
	error            error
}

func (m *mockTemplate) Execute(w io.Writer, vars any) error {
	m.executed = true
	m.lastVars = vars
	m.lastW = w
	return m.error
}

func (m *mockTemplate) ExecuteTemplate(w io.Writer, name string, vars any) error {
	m.executedTemplate = true
	m.lastVars = vars
	m.lastW = w
	return m.error
}

type mockRoute struct {
	client   ApiClient
	data     any
	executed bool
	lastW    io.Writer
	error
}

func (r *mockRoute) Client() ApiClient {
	return r.client
}

func (r *mockRoute) execute(w http.ResponseWriter, req *http.Request, data any) error {
	r.executed = true
	r.lastW = w
	r.data = data
	return r.error
}

type mockApiClient struct {
	error              error
	CurrentUserDetails shared.Assignee
	PersonDetails      shared.Person
	FeeReductions      shared.FeeReductions
	Invoices           shared.Invoices
	AccountInformation shared.AccountInformation
	invoiceAdjustments shared.InvoiceAdjustments
}

func (m mockApiClient) CancelFeeReduction(context api.Context, i int, i2 int, s string) error {
	return m.error
}

func (m mockApiClient) UpdateInvoice(api.Context, int, int, string, string, string) error {
	return m.error
}

func (m mockApiClient) AddFeeReduction(api.Context, int, string, string, string, string, string) error {
	return m.error
}

func (m mockApiClient) GetInvoices(api.Context, int) (shared.Invoices, error) {
	return m.Invoices, m.error
}

func (m mockApiClient) GetPersonDetails(api.Context, int) (shared.Person, error) {
	return m.PersonDetails, m.error
}

func (m mockApiClient) GetCurrentUserDetails(api.Context) (shared.Assignee, error) {
	return m.CurrentUserDetails, m.error
}

func (m mockApiClient) GetFeeReductions(api.Context, int) (shared.FeeReductions, error) {
	return m.FeeReductions, m.error
}

func (m mockApiClient) GetAccountInformation(api.Context, int) (shared.AccountInformation, error) {
	return m.AccountInformation, m.error
}

func (m mockApiClient) GetInvoiceAdjustments(api.Context, int) (shared.InvoiceAdjustments, error) {
	return m.invoiceAdjustments, m.error
}
