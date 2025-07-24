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

func (c *Client) AddDirectDebit(ctx context.Context, clientId int, accountName string, sortCode string, accountNumber string) error {
	var body bytes.Buffer

	// fetch client from Sirius
	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		return err
	}

	mandate := shared.NewCreateMandate(client, accountName, sortCode, accountNumber)

	// TODO: Validate

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

func (c *Client) validateMandate(client shared.Person, accountName string, sortCode string, accountNumber string) (*shared.CreateMandate, *apierror.ValidationErrors) {
	var (
		errors apierror.ValidationErrors // map[string]map[string]string
	)

	if client.FeePayer == nil {
		errors["FeePayer"] = map[string]string{
			"required": "",
		}
	} else if client.FeePayer.Status != "Active" {
		errors["FeePayer"] = map[string]string{
			"inactive": "",
		}
	}

	if client.ActiveCaseType == nil {
		errors["ActiveOrder"] = map[string]string{
			"required": "",
		}
	}

	if len(errors) > 0 {
		return nil, &errors
	}

	if accountName == "" {
		errors["AccountName"] = map[string]string{
			"required": "",
		}
	} else if len(accountName) > 18 {
		errors["AccountName"] = map[string]string{
			"gteEighteen": "",
		}
	}

	if sortCode == "" {
		errors["SortCode"] = map[string]string{
			"required": "",
		}
	} else if valid, _ := regexp.MatchString(`^\d{2}-\d{2}-\d{2}$`, sortCode); !valid {
		errors["SortCode"] = map[string]string{
			"len": "",
		}
	}

	if accountNumber == "" {
		errors["AccountNumber"] = map[string]string{
			"required": "",
		}
	} else if valid, _ := regexp.MatchString(`^\d{8}$`, accountName); !valid {
		errors["SortCode"] = map[string]string{
			"len": "",
		}
	}

	// TODO: Validate address field length
	// TODO: Call modulus check

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
				AccountName:   accountName,
				SortCode:      sortCode,
				AccountNumber: accountNumber,
			},
		},
	}, nil
}
