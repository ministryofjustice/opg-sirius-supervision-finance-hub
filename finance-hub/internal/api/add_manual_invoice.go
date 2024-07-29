package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) AddManualInvoice(ctx Context, clientId int, invoiceType string, amount *string, raisedDate *string, startDate *string, endDate *string, supervisionLevel string) error {
	var body bytes.Buffer
	var raisedDateTransformed shared.NillableDate
	var startDateTransformed shared.NillableDate
	var endDateTransformed shared.NillableDate

	if startDate != nil {
		fmt.Println("Request start date not null")
		startDateTransformed = shared.NillableDate{
			shared.NewDate(*startDate),
			true,
		}
	}

	if raisedDate != nil {
		raisedDateTransformed = shared.NillableDate{
			shared.NewDate(*raisedDate),
			true,
		}
	}

	if endDate != nil {
		endDateTransformed = shared.NillableDate{
			shared.NewDate(*endDate),
			true,
		}
	}

	addManualInvoiceForm := shared.AddManualInvoice{
		InvoiceType:      shared.ParseInvoiceType(invoiceType),
		RaisedDate:       raisedDateTransformed,
		StartDate:        startDateTransformed,
		EndDate:          endDateTransformed,
		SupervisionLevel: supervisionLevel,
	}

	if amount != nil {
		addManualInvoiceForm.Amount = shared.NillableInt{
			Value: shared.DecimalStringToInt(*amount),
			Valid: true,
		}
	}

	err := json.NewEncoder(&body).Encode(addManualInvoiceForm)

	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/invoices", clientId)
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
		var badRequests shared.BadRequests
		if err := json.NewDecoder(resp.Body).Decode(&badRequests); err != nil {
			return err
		}

		validationErrors := make(shared.ValidationErrors)
		for _, reason := range badRequests.Reasons {
			innerMap := make(map[string]string)
			innerMap[reason] = reason
			validationErrors[reason] = innerMap
		}

		return shared.ValidationError{Errors: validationErrors}
	}

	return newStatusError(resp)
}
