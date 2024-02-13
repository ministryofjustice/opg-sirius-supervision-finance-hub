package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type mockFinanceVarsClient struct {
	count      map[string]int
	lastCtx    sirius.Context
	err        error
	userData   model.Assignee
	personData model.Person
}

func (m *mockFinanceVarsClient) GetPersonDetails(ctx sirius.Context, i int) (model.Person, error) {
	if m.count == nil {
		m.count = make(map[string]int)
	}
	m.count["GetPersonDetails"] += 1
	m.lastCtx = ctx

	return m.personData, m.err
}

func (m *mockFinanceVarsClient) GetCurrentUserDetails(ctx sirius.Context) (model.Assignee, error) {
	if m.count == nil {
		m.count = make(map[string]int)
	}
	m.count["GetCurrentUserDetails"] += 1
	m.lastCtx = ctx

	return m.userData, m.err
}

var mockUserDetailsData = model.Assignee{
	Id: 123,
}

var mockPerson = model.Person{
	ID:                 1,
	Firstname:          "Person",
	Surname:            "LastName",
	CourtRef:           "12345678",
	OutstandingBalance: "£554",
	CreditBalance:      "£0.20",
	PaymentMethod:      "Demanded",
}

func TestNewFinanceVars(t *testing.T) {
	client := &mockFinanceVarsClient{userData: mockUserDetailsData, personData: mockPerson}
	r, _ := http.NewRequest("GET", "/path/1", nil)

	envVars := EnvironmentVars{}
	vars, err := NewFinanceVars(client, r, envVars)

	assert.Nil(t, err)
	assert.Equal(t, FinanceVars{
		Path:            "/path/1",
		XSRFToken:       "",
		MyDetails:       mockUserDetailsData,
		Person:          mockPerson,
		Tabs:            []Tab{{Title: "Invoices", BasePath: "invoices"}},
		SuccessMessage:  "",
		Errors:          nil,
		EnvironmentVars: envVars,
	}, *vars)
}

func TestTab_GetURL(t *testing.T) {
	tab := Tab{BasePath: "test-path"}
	personId := mockPerson.ID

	assert.Equal(t, "/1", tab.GetURL(personId))
}
