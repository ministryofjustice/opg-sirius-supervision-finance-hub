package allpay

import (
	"encoding/json"
	"errors"
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestCreateSchedule_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/AllpayApi/Customers/SCHEME123/REF123/Doe/VariableMandates" {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var req createScheduleRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("Invalid JSON body: %v", err)
		}
		expected := createScheduleRequest{
			Schedules: []schedule{
				{
					Date:          time.Time{}.Format("2006-01-02"),
					Amount:        12345,
					Frequency:     "1",
					TotalPayments: 1,
				},
			},
		}
		assert.Equal(t, req, expected)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: schemeCode,
			apiHost:    ts.URL,
		},
	}

	err := c.CreateSchedule(testContext(), CreateScheduleInput{
		ClientRef: "REF123",
		Surname:   "Doe",
		Date:      time.Time{},
		Amount:    12345,
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreateSchedule_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.CreateSchedule(testContext(), CreateScheduleInput{
		ClientRef: "REF123",
		Surname:   "Doe",
		Date:      time.Time{},
		Amount:    12345,
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestCreateSchedule_UnexpectedStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    ts.URL,
		},
	}

	err := c.CreateSchedule(testContext(), CreateScheduleInput{
		ClientRef: "REF123",
		Surname:   "Doe",
		Date:      time.Time{},
		Amount:    12345,
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}

func TestCreateSchedule_ValidationErrorValidJSON(t *testing.T) {
	ve := ErrorValidation{Messages: []string{"Invalid data"}}
	body, _ := json.Marshal(ve)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(body)
	}))
	defer ts.Close()

	c := &Client{
		http: ts.Client(),
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    ts.URL,
		},
	}

	err := c.CreateSchedule(testContext(), CreateScheduleInput{
		ClientRef: "REF123",
		Surname:   "Doe",
		Date:      time.Time{},
		Amount:    12345,
	})
	var validationErr ErrorValidation
	if !errors.As(err, &validationErr) {
		t.Errorf("Expected ErrorValidation, got %v", err)
	}
}
