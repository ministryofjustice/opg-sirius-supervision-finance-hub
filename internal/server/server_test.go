package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"io"
)

type mockTemplate struct {
	count    int
	lastVars interface{}
	lastW    io.Writer
	error    error
}

func (m *mockTemplate) Execute(w io.Writer, vars any) error {
	m.count += 1
	m.lastVars = vars
	m.lastW = w
	return m.error
}

type mockApiClient struct {
	error              error
	CurrentUserDetails model.Assignee
	PersonDetails      model.Person
}

func (m mockApiClient) GetPersonDetails(context sirius.Context, i int) (model.Person, error) {
	return m.PersonDetails, m.error
}

func (m mockApiClient) GetCurrentUserDetails(context sirius.Context) (model.Assignee, error) {
	return m.CurrentUserDetails, m.error
}
