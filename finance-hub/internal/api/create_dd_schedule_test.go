package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateDirectDebitSchedule(t *testing.T) {
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
}

// TODO: Add error cases
