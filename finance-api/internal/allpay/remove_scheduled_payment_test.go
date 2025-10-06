package allpay

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRemoveScheduledPayment_Success(t *testing.T) {
	schemeCode := "SCHEME123"

	date := time.Now()
	amount := int32(12345)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != fmt.Sprintf("/AllpayApi/Customers/SCHEME123/MTIzNDU2Nzg=/Q2FuY2VsbWFu/Mandates/Schedule/%s/%d", date.Format("2006-01-02"), amount) {
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

	err := c.RemoveScheduledPayment(testContext(), &RemoveScheduledPaymentRequest{
		CollectionDate: date,
		Amount:         amount,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestRemoveScheduledPayment_RequestCreationFails(t *testing.T) {
	c := &Client{
		http: http.DefaultClient,
		Envs: Envs{
			schemeCode: "SCHEME123",
			apiHost:    "http://[::1]:namedport/", // Invalid URL
		},
	}

	err := c.RemoveScheduledPayment(testContext(), &RemoveScheduledPaymentRequest{
		CollectionDate: time.Now(),
		Amount:         12345,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})
	if err == nil {
		t.Error("Expected error due to request creation failure")
	}
}

func TestRemoveScheduledPayment_UnexpectedStatus(t *testing.T) {
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

	err := c.RemoveScheduledPayment(testContext(), &RemoveScheduledPaymentRequest{
		CollectionDate: time.Now(),
		Amount:         12345,
		ClientDetails: ClientDetails{
			ClientReference: "12345678",
			Surname:         "Cancelman",
		},
	})
	if err == nil {
		t.Error("Expected error due to unexpected status code")
	}
}
