package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) AddManualInvoice(ctx Context, clientId int, invoiceType string, amount *string, raisedDate *string, raisedYear *string, startDate *string, endDate *string, supervisionLevel string) error {
	var body bytes.Buffer

	addManualInvoiceForm := shared.AddManualInvoice{
		InvoiceType:      shared.ParseInvoiceType(invoiceType),
		Amount:           shared.TransformNillableInt(amount),
		RaisedDate:       shared.TransformNillableDate(raisedDate),
		RaisedYear:       shared.TransformNillableInt(raisedYear),
		StartDate:        shared.TransformNillableDate(startDate),
		EndDate:          shared.TransformNillableDate(endDate),
		SupervisionLevel: supervisionLevel,
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
