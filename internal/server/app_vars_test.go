package server

import (
	"github.com/opg-sirius-finance-hub/internal/model"
	"github.com/opg-sirius-finance-hub/internal/sirius"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

type mockFinanceVarsClient struct {
	count    map[string]int
	lastCtx  sirius.Context
	err      error
	userData model.Assignee
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

func TestNewFinanceVars(t *testing.T) {
	client := &mockFinanceVarsClient{userData: mockUserDetailsData}
	r, _ := http.NewRequest("GET", "/path", nil)

	envVars := EnvironmentVars{}
	vars, err := NewAppVars(client, r, envVars)

	assert.Nil(t, err)
	assert.Equal(t, AppVars{
		Path:            "/path",
		XSRFToken:       "",
		MyDetails:       mockUserDetailsData,
		SuccessMessage:  "",
		Errors:          nil,
		EnvironmentVars: envVars,
	}, *vars)
}
