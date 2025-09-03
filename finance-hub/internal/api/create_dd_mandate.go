package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type AccountDetails struct {
	AccountName   string
	AccountNumber string
	SortCode      string
}

func (c *Client) CreateDirectDebitMandate(ctx context.Context, clientId int, details AccountDetails) error {
	var body bytes.Buffer

	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		return err
	}

	errs := c.validateActiveClient(client)
	if len(errs) > 0 {
		return apierror.ValidationError{Errors: errs}
	}

	err = json.NewEncoder(&body).Encode(shared.CreateMandate{
		ClientReference: client.CourtRef,
		Surname:         client.Surname,
		Address: shared.Address{
			Line1:    client.AddressLine1,
			Town:     client.Town,
			PostCode: client.PostCode,
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"bankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   details.AccountName,
				SortCode:      details.SortCode,
				AccountNumber: details.AccountNumber,
			},
		},
	})
	if err != nil {
		return err
	}

	req, err := c.newBackendRequest(ctx, http.MethodPost, fmt.Sprintf("/clients/%d/direct-debit", clientId), &body)

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
		return newStatusError(resp)
	}

	if resp.StatusCode == http.StatusBadRequest {
		// TODO: Mod validation here
		return apierror.ValidationError{Errors: apierror.ValidationErrors{"AccountDetails": {"invalid": ""}}}
	}

	return newStatusError(resp)
}

func (c *Client) validateActiveClient(client shared.Person) apierror.ValidationErrors {
	vErrs := make(apierror.ValidationErrors) // map[string]map[string]string

	if client.FeePayer == nil || client.FeePayer.Status != "Active" {
		vErrs["FeePayer"] = map[string]string{
			"inactive": "",
		}
	}

	if client.ActiveCaseType == nil {
		vErrs["ActiveOrder"] = map[string]string{
			"required": "",
		}
	}

	if client.ClientStatus == nil || client.ClientStatus.Handle != "ACTIVE" {
		vErrs["ClientStatus"] = map[string]string{
			"inactive": "",
		}
	}

	return vErrs
}
