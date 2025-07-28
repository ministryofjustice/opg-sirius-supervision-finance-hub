package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"net/http"
	"regexp"
)

type AccountDetails struct {
	AccountName   string
	AccountNumber string
	SortCode      string
}

func (c *Client) AddDirectDebit(ctx context.Context, clientId int, details AccountDetails) error {
	var body bytes.Buffer

	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		return err
	}

	mandate, errors := c.validateMandate(client, details)

	if errors != nil {
		return apierror.ValidationError{Errors: *errors}
	}

	err = json.NewEncoder(&body).Encode(mandate)
	if err != nil {
		return err
	}

	// TODO: Send to AllPay
	url := fmt.Sprintf("/clients/%d/direct-debits", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodPost, url, &body)

	// TODO: On success, update payment method

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

	return newStatusError(resp)
}

func (c *Client) validateBankDetails(details AccountDetails) *apierror.ValidationErrors {
	var errors apierror.ValidationErrors // map[string]map[string]string

	c.newAllPayRequest(ctx, http.MethodGet, fmt.Sprintf("/BankAccounts?sortcode=%s&accountnumber=%s", details.SortCode, details.AccountNumber), nil)

	return &errors
}

func (c *Client) validateMandate(client shared.Person, details AccountDetails) (*shared.CreateMandate, *apierror.ValidationErrors) {
	var errors apierror.ValidationErrors // map[string]map[string]string

	if client.FeePayer == nil || client.FeePayer.Status != "Active" {
		errors["FeePayer"] = map[string]string{
			"inactive": "",
		}
	}

	if client.ActiveCaseType == nil {
		errors["ActiveOrder"] = map[string]string{
			"required": "",
		}
	}

	if client.ClientStatus == nil || client.ClientStatus.Handle != "ACTIVE" {
		errors["ClientStatus"] = map[string]string{
			"inactive": "",
		}
	}

	if len(errors) > 0 {
		return nil, &errors
	}

	if details.AccountName == "" {
		errors["AccountName"] = map[string]string{
			"required": "",
		}
	} else if len(details.AccountName) > 18 {
		errors["AccountName"] = map[string]string{
			"gteEighteen": "",
		}
	}

	if details.SortCode == "" {
		errors["SortCode"] = map[string]string{
			"required": "",
		}
	} else if valid, _ := regexp.MatchString(`^\d{2}-\d{2}-\d{2}$`, details.SortCode); !valid {
		errors["SortCode"] = map[string]string{
			"len": "",
		}
	}

	if details.AccountNumber == "" {
		errors["AccountNumber"] = map[string]string{
			"required": "",
		}
	} else if valid, _ := regexp.MatchString(`^\d{8}$`, details.AccountName); !valid {
		errors["SortCode"] = map[string]string{
			"len": "",
		}
	}

	if assertStringLessThan(client.AddressLine1, 41) {
		errors["AddressLine1"] = map[string]string{
			"required": "",
		}
	}
	if assertStringLessThan(client.Town, 41) {
		errors["Town"] = map[string]string{
			"required": "",
		}
	}
	if assertStringLessThan(client.PostCode, 11) {
		errors["PostCode"] = map[string]string{
			"required": "",
		}
	}

	if len(errors) > 0 {
		return nil, &errors
	}

	return &shared.CreateMandate{
		SchemeCode:      c.SchemeCode,
		ClientReference: client.CourtRef,
		Surname:         client.Surname,
		Address: shared.Address{
			Line1:    client.AddressLine1,
			Town:     client.Town,
			PostCode: client.PostCode,
		},
		BankAccount: struct {
			BankDetails shared.AllPayBankDetails `json:"BankDetails"`
		}{
			BankDetails: shared.AllPayBankDetails{
				AccountName:   details.AccountName,
				SortCode:      details.SortCode,
				AccountNumber: details.AccountNumber,
			},
		},
	}, nil
}

func assertStringLessThan(str string, n int) bool {
	return len(str) > 0 && len(str) < n
}
