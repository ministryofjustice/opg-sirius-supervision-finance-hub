package api

import (
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateDirectDebitSchedule(t *testing.T) {
	var pendingCollectionsCalled bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111"
			}`))
		case "/clients/1/balance/pending":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`10000`))
		case "/clients/1/pending-collections":
			pendingCollectionsCalled = true
			w.WriteHeader(http.StatusCreated)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	ctx := testContext()
	err := client.CreateDirectDebitSchedule(ctx, 1)
	assert.Equal(t, nil, err)

	assert.True(t, mockAllPay.createScheduleCalled)

	date, _ := client.addWorkingDays(ctx, time.Now(), 14)
	date, _ = client.lastWorkingDayOfMonth(ctx, date)

	expected := allpay.CreateScheduleInput{
		ClientRef: "11111111",
		Surname:   "Holder",
		Date:      date,
		Amount:    10000,
	}
	assert.Equal(t, expected, mockAllPay.data.(allpay.CreateScheduleInput))
	assert.True(t, pendingCollectionsCalled)
}

func TestCreateDirectDebitSchedule_GetPendingOutstandingBalanceFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111"
			}`))
		case "/clients/1/balance/pending":
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	ctx := testContext()
	err := client.CreateDirectDebitSchedule(ctx, 1)

	assert.Error(t, err)
}

func TestCreateDirectDebitSchedule_GetPersonDetailsFails(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusInternalServerError)
		case "/clients/1/balance/pending":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`10000`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	ctx := testContext()
	err := client.CreateDirectDebitSchedule(ctx, 1)

	assert.Error(t, err)
}

func TestCreateDirectDebitSchedule_AddWorkingDaysFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111"
			}`))
		case "/clients/1/balance/pending":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`10000`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)
	client.caches.holidays = cache.New(defaultExpiration, defaultExpiration) // set to empty cache so it fails when trying to refresh

	ctx := testContext()
	err := client.CreateDirectDebitSchedule(ctx, 1)

	assert.Error(t, err)
}

func TestCreateDirectDebitSchedule_CreateScheduleFailed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111"
			}`))
		case "/clients/1/balance/pending":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`10000`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{
		createScheduleError: fmt.Errorf("createScheduleError"),
	}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	ctx := testContext()
	err := client.CreateDirectDebitSchedule(ctx, 1)

	assert.Error(t, err)
	assert.Equal(t, "createScheduleError", err.Error())
}

func TestCreateDirectDebitSchedule_pendingCollectionsFailed(t *testing.T) {
	var pendingCollectionsCalled bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/clients/1":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
			  "id": 1,
			  "firstname": "Account",
			  "surname": "Holder",
			  "caseRecNumber": "11111111"
			}`))
		case "/clients/1/balance/pending":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`10000`))
		case "/clients/1/pending-collections":
			pendingCollectionsCalled = true
			w.WriteHeader(http.StatusInternalServerError)
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	mockAllPay := mockAllPayClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, BackendURL: ts.URL}, &mockAllPay)

	logHandler := TestLogHandler{}
	err := client.CreateDirectDebitSchedule(testContextWithLogger(&logHandler), 1)

	assert.True(t, pendingCollectionsCalled)
	assert.Error(t, err)
	logHandler.assertLog(t, "failed to create pending collection in Sirius after successful schedule instruction in AllPay")
}
