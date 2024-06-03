package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) AdjustInvoice(ctx Context, clientId int, supervisionBillingTeamId int, invoiceId int, adjustmentType string, notes string, amount string) error {
	var body bytes.Buffer

	adjustment := shared.CreateLedgerEntryRequest{
		AdjustmentType:  shared.ParseAdjustmentType(adjustmentType),
		AdjustmentNotes: notes,
		Amount:          shared.DecimalStringToInt(amount),
	}

	err := json.NewEncoder(&body).Encode(adjustment)
	if err != nil {
		return err
	}

	req, err := c.newBackendRequest(ctx, http.MethodPost, fmt.Sprintf("/clients/%d/invoices/%d/ledger-entries", clientId, invoiceId), &body)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var invoiceRef string

	if err = json.NewDecoder(resp.Body).Decode(&invoiceRef); err != nil {
		c.logger.Request(req, err)
		return err
	}

	if resp.StatusCode == http.StatusCreated {
		err := c.CreatePendingInvoiceAdjustmentTask(ctx, clientId, supervisionBillingTeamId, invoiceRef, adjustmentType)
		if err != nil {
			return err
		}

		return nil
	}
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}
	if resp.StatusCode == http.StatusUnprocessableEntity {
		var v shared.ValidationError
		if err = json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return shared.ValidationError{Errors: v.Errors}
		}
	}
	if resp.StatusCode == http.StatusBadRequest {
		var be shared.BadRequest
		if err = json.NewDecoder(resp.Body).Decode(&be); err == nil {
			return be
		}
	}

	return newStatusError(resp)
}
