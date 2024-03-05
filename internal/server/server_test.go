package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
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
	CurrentUserDetails model.Assignee
	PersonDetails      model.Person
	FeeReductions      model.FeeReductions
	Invoices           model.Invoices
}

func (m mockApiClient) UpdateInvoice(context sirius.Context, i int, k int, s string, j string, amount string) error {
	return m.error
}

func (m mockApiClient) GetInvoices(context sirius.Context, i int) (model.Invoices, error) {
	return m.Invoices, m.error
}

func (m mockApiClient) GetPersonDetails(sirius.Context, int) (model.Person, error) {
	return m.PersonDetails, m.error
}

func (m mockApiClient) GetCurrentUserDetails(sirius.Context) (model.Assignee, error) {
	return m.CurrentUserDetails, m.error
}

func (m mockApiClient) GetFeeReductions(sirius.Context, int) (model.FeeReductions, error) {
	return m.FeeReductions, m.error
}
