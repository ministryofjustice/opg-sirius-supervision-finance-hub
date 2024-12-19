package api

import (
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"strconv"
)

func (c *ApiClient) SubmitDirectDebit(accountHolder string, accountName string, sortCode string, accountNumber string) error {
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

	checkSortCodeIsNotJustZeros, _ := strconv.Atoi(sortCode)

	if checkSortCodeIsNotJustZeros == 0 || len(sortCode) != 6 {
		errors["SortCode"] = map[string]string{"eqSix": ""}
	}

	if len(accountNumber) != 8 {
		errors["AccountNumber"] = map[string]string{"eqEight": ""}
	}

	if len(errors) > 0 {
		return apierror.ValidationError{
			Errors: errors,
		}
	}

	return nil
}
