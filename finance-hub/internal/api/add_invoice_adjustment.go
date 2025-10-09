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

func (c *Client) AddInvoiceAdjustment(ctx context.Context, clientId int, supervisionBillingTeamId int, invoiceId int, adjustmentType string, notes string, amount string, managerOverride bool) error {
	var body bytes.Buffer

	adjustment := shared.AddInvoiceAdjustmentRequest{
		AdjustmentType:  shared.ParseAdjustmentType(adjustmentType),
		AdjustmentNotes: notes,
		Amount:          shared.DecimalStringToInt(amount),
		ManagerOverride: managerOverride,
	}

	err := json.NewEncoder(&body).Encode(adjustment)
	if err != nil {
		return err
	}

	req, err := c.newBackendRequest(ctx, http.MethodPost, fmt.Sprintf("/clients/%d/invoices/%d/invoice-adjustments", clientId, invoiceId), &body)

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
		var response shared.InvoiceReference
		if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return err
		}

		err := c.CreatePendingInvoiceAdjustmentTask(ctx, clientId, supervisionBillingTeamId, response.InvoiceRef, adjustmentType)
		if err != nil {
			return err
		}

		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var v apierror.ValidationError
		if err = json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return apierror.ValidationError{Errors: v.Errors}
		}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var be apierror.BadRequest
		if err = json.NewDecoder(resp.Body).Decode(&be); err == nil {
			return be
		}
	}

	return newStatusError(resp)
}
