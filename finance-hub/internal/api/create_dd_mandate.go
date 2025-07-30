package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/ministryofjustice/opg-go-common/telemetry"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/finance-hub/internal/allpay"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"regexp"
)

type AccountDetails struct {
	AccountName   string
	AccountNumber string
	SortCode      string
}

func (c *Client) CreateDirectDebitMandate(ctx context.Context, clientId int, details AccountDetails) error {
	var body bytes.Buffer
	logger := telemetry.LoggerFromContext(ctx)

	client, err := c.GetPersonDetails(ctx, clientId)
	if err != nil {
		return err
	}

	mandate, errs := c.validateMandate(client, details)
	if errs != nil {
		return apierror.ValidationError{Errors: *errs}
	}

	err = json.NewEncoder(&body).Encode(mandate)
	if err != nil {
		return err
	}

	err = c.allpayClient.ModulusCheck(ctx, details.SortCode, details.AccountNumber)
	if err != nil {
		if errors.As(err, &allpay.ErrorModulusCheckFailed{}) {
			return apierror.ValidationError{Errors: apierror.ValidationErrors{
				"AccountDetails": map[string]string{
					"invalid": "",
				},
			}}
		}
		return err
	}

	err = c.allpayClient.CreateMandate(ctx, mandate)
	if err != nil {
		var ve allpay.ErrorValidation
		if errors.As(err, &ve) {
			// we validate in advance so validation errors from AllPay should never occur
			// if they do, log them so we can investigate
			logger.Error("validation errors returned from allpay", "errors", ve.Messages)
		}
		return err
	}

	err = c.UpdatePaymentMethod(ctx, clientId, shared.PaymentMethodDirectDebit.Key())
	if err != nil {
		logger.Error("failed to update payment method in Sirius after successful mandate creation in AllPay", "error", err)
		return err
	}
	return nil
}

func (c *Client) validateMandate(client shared.Person, details AccountDetails) (*allpay.CreateMandateRequest, *apierror.ValidationErrors) {
	vErrs := make(apierror.ValidationErrors) // map[string]map[string]string

	if details.AccountName == "" {
		vErrs["AccountName"] = map[string]string{
			"required": "",
		}
	} else if len(details.AccountName) > 18 {
		vErrs["AccountName"] = map[string]string{
			"gteEighteen": "",
		}
	}

	if details.SortCode == "" {
		vErrs["SortCode"] = map[string]string{
			"required": "",
		}
	} else if valid, _ := regexp.MatchString(`^\d{2}-\d{2}-\d{2}$`, details.SortCode); !valid {
		vErrs["SortCode"] = map[string]string{
			"len": "",
		}
	}

	if details.AccountNumber == "" {
		vErrs["AccountNumber"] = map[string]string{
			"required": "",
		}
	} else if valid, _ := regexp.MatchString(`^\d{8}$`, details.AccountNumber); !valid {
		vErrs["AccountNumber"] = map[string]string{
			"len": "",
		}
	}

	// validate form fields first
	if len(vErrs) > 0 {
		return nil, &vErrs
	}

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

	if !assertStringLessThan(client.AddressLine1, 41) {
		vErrs["AddressLine1"] = map[string]string{
			"required": "",
		}
	}
	if !assertStringLessThan(client.Town, 41) {
		vErrs["Town"] = map[string]string{
			"required": "",
		}
	}
	if !assertStringLessThan(client.PostCode, 11) {
		vErrs["PostCode"] = map[string]string{
			"required": "",
		}
	}

	if len(vErrs) > 0 {
		return nil, &vErrs
	}

	return &allpay.CreateMandateRequest{
		ClientReference: client.CourtRef,
		Surname:         client.Surname,
		Address: allpay.Address{
			Line1:    client.AddressLine1,
			Town:     client.Town,
			PostCode: client.PostCode,
		},
		BankAccount: struct {
			BankDetails allpay.BankDetails `json:"BankDetails"`
		}{
			BankDetails: allpay.BankDetails{
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
