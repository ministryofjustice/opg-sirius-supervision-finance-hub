package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type BadRequest struct {
	Reasons []string `json:"reasons"`
}

func (c *ApiClient) AddManualInvoice(ctx Context, clientId int, invoiceType string, amount string, raisedDate string, startDate string, endDate string, supervisionLevel string) error {
	var body bytes.Buffer
	var raisedDateTransformed *shared.Date
	var startDateTransformed *shared.Date
	var endDateTransformed *shared.Date

	if raisedDate != "" {
		raisedDateToTime, _ := time.Parse("2006-01-02", raisedDate)
		raisedDateTransformed = &shared.Date{Time: raisedDateToTime}
	}

	if startDate != "" {
		startDateToTime, _ := time.Parse("2006-01-02", startDate)
		startDateTransformed = &shared.Date{Time: startDateToTime}
	}

	if endDate != "" {
		endDateToTime, _ := time.Parse("2006-01-02", endDate)
		endDateTransformed = &shared.Date{Time: endDateToTime}
	}

	err := json.NewEncoder(&body).Encode(shared.AddManualInvoice{
		InvoiceType:      invoiceType,
		Amount:           shared.DecimalStringToInt(amount),
		RaisedDate:       raisedDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: supervisionLevel,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/manual-invoice", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodPost, url, &body)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var v shared.ValidationError
		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return shared.ValidationError{Errors: v.Errors}
		}
	}

	if resp.StatusCode == http.StatusBadRequest {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		bodyString := strings.TrimRight(string(bodyBytes), "\n")

		var badRequest BadRequest
		if err := json.Unmarshal([]byte(bodyString), &badRequest); err != nil {
			return err
		}

		validationErrors := make(shared.ValidationErrors)
		for _, reason := range badRequest.Reasons {
			innerMap := make(map[string]string)
			innerMap[reason] = reason
			validationErrors[reason] = innerMap
		}

		return shared.ValidationError{Errors: validationErrors}
	}

	return newStatusError(resp)
}
