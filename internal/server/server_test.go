package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"io"
)

type mockTemplate struct {
	execute         bool
	executeTemplate bool
	lastVars        interface{}
	lastW           io.Writer
	error           error
}

func (m *mockTemplate) Execute(w io.Writer, vars any) error {
	m.execute = true
	m.lastVars = vars
	m.lastW = w
	return m.error
}

func (m *mockTemplate) ExecuteTemplate(w io.Writer, name string, vars any) error {
	m.executeTemplate = true
	m.lastVars = vars
	m.lastW = w
	return m.error
}

type mockApiClient struct {
	error              error
	CurrentUserDetails model.Assignee
	PersonDetails      model.Person
	InvoicesList       model.InvoiceList
}

func (m mockApiClient) GetInvoices(context sirius.Context, i int) (model.InvoiceList, error) {
	return m.InvoicesList, m.error
}

func (m mockApiClient) GetPersonDetails(sirius.Context, int) (model.Person, error) {
	return m.PersonDetails, m.error
}

func (m mockApiClient) GetCurrentUserDetails(sirius.Context) (model.Assignee, error) {
	return m.CurrentUserDetails, m.error
}
