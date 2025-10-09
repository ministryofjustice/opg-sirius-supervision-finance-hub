package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
)

func (c *Client) AddManualInvoice(ctx context.Context, clientId int, invoiceType string, amount *string, raisedDate *string, raisedYear *string, startDate *string, endDate *string, supervisionLevel *string) error {
	var body bytes.Buffer

	addManualInvoiceForm := shared.AddManualInvoice{
		InvoiceType:      shared.ParseInvoiceType(invoiceType),
		Amount:           shared.TransformNillableInt(amount),
		StartDate:        shared.TransformNillableDate(startDate),
		EndDate:          shared.TransformNillableDate(endDate),
		SupervisionLevel: shared.TransformNillableString(supervisionLevel),
	}

	if raisedYear != nil && *raisedYear != "" {
		raisedYearDate := *raisedYear + "-03-31"
		addManualInvoiceForm.RaisedDate = shared.TransformNillableDate(&raisedYearDate)
	} else {
		addManualInvoiceForm.RaisedDate = shared.TransformNillableDate(raisedDate)
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

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusCreated {
		return nil
	}

	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var v apierror.ValidationError
		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return apierror.ValidationError{Errors: v.Errors}
		}
	}

	if resp.StatusCode == http.StatusBadRequest {
		var badRequests apierror.BadRequests
		if err := json.NewDecoder(resp.Body).Decode(&badRequests); err != nil {
			return err
		}

		validationErrors := make(apierror.ValidationErrors)
		for _, reason := range badRequests.Reasons {
			innerMap := make(map[string]string)
			innerMap[reason] = reason
			validationErrors[reason] = innerMap
		}

		return apierror.ValidationError{Errors: validationErrors}
	}

	return newStatusError(resp)
}
