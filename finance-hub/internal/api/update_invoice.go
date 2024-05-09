package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type updateInvoice struct {
	AdjustmentType string `json:"adjustmentType"`
	Notes          string `json:"notes"`
	Amount         string `json:"amount"`
}

func (c *ApiClient) UpdateInvoice(ctx Context, clientId int, invoiceId int, adjustmentType string, notes string, amount string) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(updateInvoice{
		AdjustmentType: adjustmentType,
		Notes:          notes,
		Amount:         amount,
	})
	if err != nil {
		return err
	}

	req, err := c.newSiriusRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/invoices/%d/ledger-entries", invoiceId), &body)

	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusUnauthorized {
		return ErrUnauthorized
	}

	if resp.StatusCode != http.StatusCreated {
		var v struct {
			ValidationErrors ValidationErrors `json:"validation_errors"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.ValidationErrors) > 0 {
			return ValidationError{Errors: v.ValidationErrors}
		}

		return newStatusError(resp)
	}

	return nil
}
