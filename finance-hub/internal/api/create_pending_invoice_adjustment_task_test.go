package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAddWorkingDays(t *testing.T) {
	var (
		startDate    time.Time
		expectedDate time.Time
	)
	tests := []struct {
		name         string
		startDate    string
		expectedDate string
		workDays     int
	}{
		{
			name:         "Start on weekday, end within working week",
			startDate:    "2024-06-04",
			expectedDate: "2024-06-07",
			workDays:     3,
		},
		{
			name:         "Start on weekday, end on following working week",
			startDate:    "2024-06-03",
			expectedDate: "2024-06-11",
			workDays:     6,
		},
		{
			name:         "Start on saturday, end on following working week",
			startDate:    "2024-06-01",
			expectedDate: "2024-06-10",
			workDays:     6,
		},
		{
			name:         "Start on sunday, end on following working week",
			startDate:    "2024-06-02",
			expectedDate: "2024-06-10",
			workDays:     6,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			startDate, _ = time.Parse("2006-01-02", test.startDate)
			expectedDate, _ = time.Parse("2006-01-02", test.expectedDate)

			assert.Equal(t, expectedDate, AddWorkingDays(startDate, test.workDays))
		})
	}
}

func TestCreatePendingInvoiceAdjustmentTask(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", "", logger)

	dueDate := AddWorkingDays(time.Now(), 20).Format("02/01/2006")

	json := fmt.Sprintf(
		`{
			"personId": "2",
			"type": "FPIA",
			"dueDate": "%s",
			"assigneeId": "41",
			"description": "Pending credit memo added to 4 requires manager approval"
        }`, dueDate)

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.CreatePendingInvoiceAdjustmentTask(getContext(nil), 2, 41, "4", "CREDIT_MEMO")
	assert.Equal(t, nil, err)
}

func TestCreatePendingInvoiceAdjustmentTaskUnauthorised(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	err := client.CreatePendingInvoiceAdjustmentTask(getContext(nil), 2, 41, "4", "CREDIT_MEMO")

	assert.Equal(t, ErrUnauthorized.Error(), err.Error())
}

func TestCreatePendingInvoiceAdjustmentTaskReturns500Error(t *testing.T) {
	logger, _ := SetUpTest()
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	err := client.CreatePendingInvoiceAdjustmentTask(getContext(nil), 2, 41, "4", "CREDIT_MEMO")
	assert.Equal(t, StatusError{
		Code:   http.StatusInternalServerError,
		URL:    svr.URL + "/api/v1/tasks",
		Method: http.MethodPost,
	}, err)
}

func TestCreatePendingInvoiceAdjustmentTaskReturnsValidationError(t *testing.T) {
	logger, _ := SetUpTest()
	validationErrors := shared.ValidationError{
		Message: "Validation failed",
		Errors: map[string]map[string]string{
			"Field": {
				"Tag": "Message",
			},
		},
	}
	responseBody, _ := json.Marshal(validationErrors)
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write(responseBody)
	}))
	defer svr.Close()

	client, _ := NewApiClient(http.DefaultClient, svr.URL, svr.URL, logger)

	err := client.CreatePendingInvoiceAdjustmentTask(getContext(nil), 2, 41, "4", "CREDIT_MEMO")
	expectedError := shared.ValidationError{Message: "", Errors: shared.ValidationErrors{"Field": map[string]string{"Tag": "Message"}}}
	assert.Equal(t, expectedError, err.(shared.ValidationError))
}
