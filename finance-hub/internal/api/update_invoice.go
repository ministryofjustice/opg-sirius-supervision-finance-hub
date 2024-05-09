package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

type updateInvoice struct {
	InvoiceType string `json:"invoiceType"`
	Notes       string `json:"notes"`
	Amount      string `json:"amount"`
}

func (c *ApiClient) UpdateInvoice(ctx Context, clientId int, invoiceId int, invoiceType string, notes string, amount string) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(updateInvoice{
		InvoiceType: invoiceType,
		Notes:       notes,
		Amount:      amount,
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
		return shared.ErrUnauthorized
	}

	if resp.StatusCode != http.StatusCreated {
		var v shared.ValidationError

		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return shared.ValidationError{Errors: v.Errors}
		}

		return shared.NewStatusError(resp)
	}

	return nil
}
