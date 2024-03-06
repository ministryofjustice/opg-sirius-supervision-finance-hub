package sirius

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type updateInvoice struct {
	InvoiceType string `json:"invoiceType"`
	Notes       string `json:"notes"`
}

func (c *ApiClient) UpdateInvoice(ctx Context, clientId int, invoiceId int, invoiceType string, notes string) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(updateInvoice{
		InvoiceType: invoiceType,
		Notes:       notes,
	})
	if err != nil {
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, fmt.Sprintf("/api/v1/invoices/%d/ledger-entries", invoiceId), &body)

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
