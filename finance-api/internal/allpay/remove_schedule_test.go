package allpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRemoveSchedule_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	date := time.Now()
	amount := 12345

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		today := date.Format("2006-01-02")
		if r.URL.Path != fmt.Sprintf("/AllpayApi/Customers/SCHEME123/MTIzNDU2Nzg=/Q2FuY2VsbWFu/Mandates/Schedule/%s/%d", today, amount) {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}
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

	err := c.RemoveSchedule(testContext(), &RemoveScheduleRequest{
		ClosureDate: date,
		Amount:      amount,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         " Cancelman ", // whitespace should be stripped before encoding
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRemoveSchedule_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.RemoveSchedule(testContext(), &RemoveScheduleRequest{
		ClosureDate: time.Now(),
		Amount:      12345,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestRemoveSchedule_UnexpectedStatus(t *testing.T) {
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

	err := c.RemoveSchedule(testContext(), &RemoveScheduleRequest{
		ClosureDate: time.Now(),
		Amount:      12345,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}
