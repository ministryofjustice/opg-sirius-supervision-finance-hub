package server

import (
	"context"
	"errors"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/auth"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockRouteData struct {
	stuff string
	AppVars
}

func TestRoute_htmxRequest(t *testing.T) {
	client := mockApiClient{}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.Header.Add("HX-Request", "true")

	data := mockRouteData{
		stuff:   "abc",
		AppVars: AppVars{Path: "/path"},
	}

	sut := route{client: client, tmpl: template, partial: "test"}

	err := sut.execute(w, r, data)

	assert.Nil(t, err)
	assert.True(t, template.executedTemplate)
	assert.False(t, template.executed)

	assert.Equal(t, data, template.lastVars)
}

func TestRoute_fullPage(t *testing.T) {
	client := mockApiClient{}
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	ctx := auth.Context{
		User:    &shared.User{ID: 123},
		Context: context.Background(),
	}
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
	r.SetPathValue("clientId", "1")

	client.PersonDetails = shared.Person{
		ID:        1,
		FirstName: "Ian",
		Surname:   "Testing",
		CourtRef:  "123abc",
	}
	client.AccountInformation = shared.AccountInformation{
		OutstandingBalance: 12300,
		CreditBalance:      123,
		PaymentMethod:      "DEMANDED",
	}
	client.CurrentUserDetails = shared.User{
		ID: 123,
	}

	fetchedData := HeaderData{
		FinanceClient: FinanceClient{
			FirstName:          client.PersonDetails.FirstName,
			Surname:            client.PersonDetails.Surname,
			CourtRef:           client.PersonDetails.CourtRef,
			OutstandingBalance: "123",
			CreditBalance:      "1.23",
			PaymentMethod:      "Demanded",
			ClientId:           "1",
		},
		User: &client.CurrentUserDetails,
	}

	data := PageData{
		Data: mockRouteData{
			stuff:   "abc",
			AppVars: AppVars{Path: "/path/"},
		},
		HeaderData: fetchedData,
	}

	sut := route{client: client, tmpl: template, partial: "test"}

	err := sut.execute(w, r, data.Data)

	assert.Nil(t, err)
	assert.True(t, template.executed)
	assert.False(t, template.executedTemplate)

	assert.Equal(t, data, template.lastVars)
}

func TestRoute_error(t *testing.T) {
	client := mockApiClient{}
	client.error = errors.New("it broke")
	template := &mockTemplate{}

	w := httptest.NewRecorder()
	ctx := auth.Context{
		User:    &shared.User{ID: 123},
		Context: context.Background(),
	}
	r, _ := http.NewRequestWithContext(ctx, http.MethodGet, "", nil)
	r.SetPathValue("clientId", "abc")

	data := PageData{
		Data: mockRouteData{
			stuff:   "abc",
			AppVars: AppVars{Path: "/path/"},
		},
	}

	sut := route{client: client, tmpl: template, partial: "test"}

	err := sut.execute(w, r, data.Data)

	assert.NotNil(t, err)
	assert.Equal(t, "it broke", err.Error())
}
