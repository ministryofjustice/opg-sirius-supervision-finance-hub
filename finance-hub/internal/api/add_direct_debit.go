package api

import (
	"bytes"
	"encoding/json"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"strconv"
	"strings"
)

func (c *Client) AddDirectDebit(accountHolder string, accountName string, sortCode string, accountNumber string) error {
	var body bytes.Buffer
	errors := make(map[string]map[string]string)

	if accountHolder == "" {
		errors["AccountHolder"] = map[string]string{"required": ""}
	}

	if accountName == "" {
		errors["AccountName"] = map[string]string{"required": ""}
	}

	if len(accountName) > 18 {
		errors["AccountName"] = map[string]string{"gteEighteen": ""}
	}

	if sortCode == "" {
		errors["SortCode"] = map[string]string{"required": ""}
	} else if len(sortCode) != 8 {
		errors["SortCode"] = map[string]string{"len": ""}
	}

	var sortCodeIsAllZeros = isSortCodeAllZeros(sortCode)
	if sortCodeIsAllZeros && len(sortCode) == 8 {
		errors["SortCode"] = map[string]string{"valid": ""}
	}

	if accountNumber == "" {
		errors["AccountNumber"] = map[string]string{"required": ""}
	} else if len(accountNumber) != 8 {
		errors["AccountNumber"] = map[string]string{"len": ""}
	}

	if len(errors) > 0 {
		return apierror.ValidationError{
			Errors: errors,
		}
	}

	err := json.NewEncoder(&body).Encode(shared.AddDirectDebit{
		AccountHolder: accountHolder,
		AccountName:   accountName,
		AccountNumber: accountNumber,
		SortCode:      sortCode,
	})

	if err != nil {
		return err
	}

	return nil
}

func isSortCodeAllZeros(sortCode string) bool {
	sortCodeWithoutDashes := strings.Split(sortCode, `-`)
	total := 0
	allZeros := true
	for i := 0; i < len(sortCodeWithoutDashes); i++ {
		convertedInt, _ := strconv.Atoi(sortCodeWithoutDashes[i])
		if convertedInt != 0 {
			allZeros = false
		}
		total += i
	}
	return allZeros
}
