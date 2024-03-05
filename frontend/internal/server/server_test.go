package server

import (
	"github.com/opg-sirius-finance-hub/api"
	"github.com/opg-sirius-finance-hub/frontend/internal/sirius"
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
	CurrentUserDetails api.Assignee
	PersonDetails      api.Person
	FeeReductions      api.FeeReductions
}

func (m mockApiClient) GetPersonDetails(sirius.Context, int) (api.Person, error) {
	return m.PersonDetails, m.error
}

func (m mockApiClient) GetCurrentUserDetails(sirius.Context) (api.Assignee, error) {
	return m.CurrentUserDetails, m.error
}

func (m mockApiClient) GetFeeReductions(sirius.Context, int) (api.FeeReductions, error) {
	return m.FeeReductions, m.error
}
