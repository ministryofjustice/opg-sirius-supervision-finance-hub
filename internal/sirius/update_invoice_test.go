package sirius

import (
	"bytes"
	"github.com/opg-sirius-finance-hub/internal/mocks"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateInvoice(t *testing.T) {
	logger, mockClient := SetUpTest()
	client, _ := NewApiClient(mockClient, "http://localhost:3000", logger)

	json := `{
            "invoiceType": "writeOff",
            "notes": "notes here",
			"amount": "100"
        }`

	r := io.NopCloser(bytes.NewReader([]byte(json)))

	mocks.GetDoFunc = func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 201,
			Body:       r,
		}, nil
	}

	err := client.UpdateInvoice(getContext(nil), 2, 4, "writeOff", "notes here", "100")
	assert.Equal(t, nil, err)
}
