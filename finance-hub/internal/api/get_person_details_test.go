package api

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"github.com/pact-foundation/pact-go/v2/consumer"
	"github.com/pact-foundation/pact-go/v2/matchers"
	"github.com/stretchr/testify/assert"
)

func TestGetPersonDetails(t *testing.T) {
	mockClient := SetUpTest()
	mockJWT := mockJWTClient{}
	client := NewClient(mockClient, &mockJWT, Envs{"http://localhost:3000", ""})

	json := `{
            "id": 2,
            "firstname": "Finance",
            "surname": "Person",
			"caseRecNumber": "12345678"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       r,
		}, nil
	}

	expectedResponse := shared.Person{
		ID:        2,
		FirstName: "Finance",
		Surname:   "Person",
		CourtRef:  "12345678",
	}

	person, err := client.GetPersonDetails(testContext(), 2)
	assert.Equal(t, expectedResponse, person)
	assert.Equal(t, nil, err)
}

func TestGetPersonDetailsReturnsUnauthorisedClientError(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})
	_, err := client.GetPersonDetails(testContext(), 2)
	assert.Equal(t, ErrUnauthorized, err)
}

func TestPersonDetailsReturns500Error(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{svr.URL, svr.URL})

	_, err := client.GetPersonDetails(testContext(), 1)
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/supervision-api/v1/clients/1",
		Method: http.MethodGet,
	}, err)
}

func TestGetPersonDetails_contract(t *testing.T) {
	pact, err := consumer.NewV2Pact(consumer.MockHTTPProviderConfig{
		Consumer: "sirius-supervision-finance-hub",
		Provider: "sirius",
		LogDir:   "../../../logs",
		PactDir:  "../../../pacts",
	})
	assert.NoError(t, err)

	err = pact.
		AddInteraction().
		UponReceiving("A request for client with ID 123").
		WithRequestPathMatcher("GET", matchers.Regex("/supervision-api/v1/clients/123", `\/supervision-api\/v1\/clients\/\d+`),
			func(b *consumer.V2RequestBuilder) {
				b.Header("Accept", matchers.S("application/json"))
			}).
		WillRespondWith(200, func(b *consumer.V2ResponseBuilder) {
			b.Header("Content-Type", matchers.S("application/json"))
			b.JSONBody(matchers.MapMatcher{
				"id":            matchers.Like(123),
				"firstname":     matchers.Like("Ian"),
				"surname":       matchers.Like("Finance"),
				"caseRecNumber": matchers.Like("11223344"),
			})
		}).
		ExecuteTest(t, func(config consumer.MockServerConfig) error {
			client := NewClient(http.DefaultClient, &mockJWTClient{}, Envs{fmt.Sprintf("http://%s:%d", config.Host, config.Port), ""})

			person, err := client.GetPersonDetails(testContext(), 123)
			if err != nil {
				return err
			}

			assert.EqualValues(t, shared.Person{
				ID:        123,
				FirstName: "Ian",
				Surname:   "Finance",
				CourtRef:  "11223344",
			}, person)

			return nil
		})

	assert.NoError(t, err)
}
