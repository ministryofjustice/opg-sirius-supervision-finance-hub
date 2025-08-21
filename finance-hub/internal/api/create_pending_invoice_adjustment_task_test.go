package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"

	"github.com/stretchr/testify/assert"
)

func TestCreatePendingInvoiceAdjustmentTask(t *testing.T) {
	var payload shared.Task
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/tasks":
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &payload); err != nil {
				t.Errorf("Invalid JSON body: %v", err)
			}
			w.WriteHeader(http.StatusCreated)
		case "/holidays.json":
			_, _ = w.Write([]byte(`{
			  "england-and-wales": {
				"division": "england-and-wales",
				"events": [
				  {
					"title": "New Year’s Day",
					"date": "2024-01-01",
					"notes": "",
					"bunting": true
				  }
				]
			  }
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, HolidayAPIURL: ts.URL + "/holidays.json"}, nil)

	dueDate, _ := client.addWorkingDays(testContext(), time.Now(), 20)
	err := client.CreatePendingInvoiceAdjustmentTask(testContext(), 2, 41, "4", "CREDIT_MEMO")
	assert.Equal(t, nil, err)

	expected := shared.Task{
		ClientId: 2,
		Type:     "FPIA",
		DueDate:  shared.Date{Time: dueDate.UTC().Truncate(24 * time.Hour)},
		Assignee: 41,
		Notes:    "Pending credit memo added to 4 requires manager approval",
	}
	assert.Equal(t, expected, payload)
}

func TestCreatePendingInvoiceAdjustmentTaskUnauthorised(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/tasks":
			w.WriteHeader(http.StatusUnauthorized)
		case "/holidays.json":
			_, _ = w.Write([]byte(`{
			  "england-and-wales": {
				"division": "england-and-wales",
				"events": [
				  {
					"title": "New Year’s Day",
					"date": "2024-01-01",
					"notes": "",
					"bunting": true
				  }
				]
			  }
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, HolidayAPIURL: ts.URL + "/holidays.json"}, nil)

	err := client.CreatePendingInvoiceAdjustmentTask(testContext(), 2, 41, "4", "CREDIT_MEMO")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestCreatePendingInvoiceAdjustmentTaskReturns500Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/tasks":
			w.WriteHeader(http.StatusInternalServerError)
		case "/holidays.json":
			_, _ = w.Write([]byte(`{
			  "england-and-wales": {
				"division": "england-and-wales",
				"events": [
				  {
					"title": "New Year’s Day",
					"date": "2024-01-01",
					"notes": "",
					"bunting": true
				  }
				]
			  }
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, HolidayAPIURL: ts.URL + "/holidays.json"}, nil)

	err := client.CreatePendingInvoiceAdjustmentTask(testContext(), 2, 41, "4", "CREDIT_MEMO")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    ts.URL + "/supervision-api/v1/tasks",
		Method: http.MethodPost,
	}, err)
}

func TestCreatePendingInvoiceAdjustmentTaskReturnsValidationError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/supervision-api/v1/tasks":
			w.WriteHeader(http.StatusUnprocessableEntity)
			validationErrors := apierror.ValidationError{
				Errors: map[string]map[string]string{
					"Field": {
						"Tag": "Message",
					},
				},
			}
			body, _ := json.Marshal(validationErrors)
			_, _ = w.Write(body)
		case "/holidays.json":
			_, _ = w.Write([]byte(`{
			  "england-and-wales": {
				"division": "england-and-wales",
				"events": [
				  {
					"title": "New Year’s Day",
					"date": "2024-01-01",
					"notes": "",
					"bunting": true
				  }
				]
			  }
			}`))
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer ts.Close()

	mockJWT := mockJWTClient{}
	client := NewClient(ts.Client(), &mockJWT, Envs{SiriusURL: ts.URL, HolidayAPIURL: ts.URL + "/holidays.json"}, nil)

	err := client.CreatePendingInvoiceAdjustmentTask(testContext(), 2, 41, "4", "CREDIT_MEMO")
	expectedError := apierror.ValidationError{Errors: apierror.ValidationErrors{"Field": map[string]string{"Tag": "Message"}}}
	assert.Equal(t, expectedError, err.(apierror.ValidationError))
}
