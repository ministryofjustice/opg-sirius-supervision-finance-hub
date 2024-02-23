package server

import (
	"errors"
	"github.com/opg-sirius-finance-hub/internal/model"
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
	r, _ := http.NewRequest(http.MethodGet, "", nil)
	r.SetPathValue("id", "1")

	fetchedData := HeaderData{
		Person: model.Person{
			Firstname: "Ian",
			Surname:   "Testing",
		},
		MyDetails: model.Assignee{
			Id: 123,
		},
	}

	client.PersonDetails = fetchedData.Person
	client.CurrentUserDetails = fetchedData.MyDetails

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
	r.SetPathValue("id", "abc")

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
