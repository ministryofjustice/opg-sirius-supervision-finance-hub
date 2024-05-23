package server

import (
	"errors"
	"github.com/opg-sirius-finance-hub/shared"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	r, _ := http.NewRequest(http.MethodGet, "", nil)
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
	client.CurrentUserDetails = shared.Assignee{
		Id: 123,
	}

	fetchedData := HeaderData{
		FinanceClient: FinanceClient{
			FirstName:          client.PersonDetails.FirstName,
			Surname:            client.PersonDetails.Surname,
			CourtRef:           client.PersonDetails.CourtRef,
			OutstandingBalance: "123",
			CreditBalance:      "1.23",
			PaymentMethod:      "Demanded",
		},
		MyDetails: client.CurrentUserDetails,
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
	r, _ := http.NewRequest(http.MethodGet, "", nil)
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

func TestRoute_GetSuccess(t *testing.T) {
	testCases := []struct {
		queryValue string
		expected   string
	}{
		{"CREDIT_WRITE_OFF", "The write off is now waiting for approval"},
		{"remission", "The remission has been successfully added"},
		{"exemption", "The exemption has been successfully added"},
		{"hardship", "The hardship has been successfully added"},
		{"credit approved", "You have approved the credit"},
		{"credit rejected", "You have rejected the credit"},
		{"write off approved", "You have approved the write off"},
		{"write off rejected", "You have rejected the write off"},
		{"invalid", ""},
	}

	req := &http.Request{
		URL: &url.URL{
			RawQuery: "",
		},
	}

	for _, tc := range testCases {
		req.URL.RawQuery = "success=" + tc.queryValue
		r := route{}
		result := r.getSuccess(req)

		if result != tc.expected {
			t.Errorf("For query value %s, expected %s, but got %s", tc.queryValue, tc.expected, result)
		}
	}
}
