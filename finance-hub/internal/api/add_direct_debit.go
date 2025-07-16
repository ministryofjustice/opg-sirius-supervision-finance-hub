package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
	"strconv"
)

func (c *Client) AddDirectDebit(ctx context.Context, clientId int, accountHolder string, accountName string, sortCode string, accountNumber string) error {
	var body bytes.Buffer
	errors := make(map[string]map[string]string)

	fmt.Println(fmt.Sprintf("Add direct debit %s, %s, %s, %s", accountHolder, accountName, sortCode, accountNumber))

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

	err := json.NewEncoder(&body).Encode(shared.AddDirectDebit{
		AccountHolder: accountHolder,
		AccountName:   accountName,
		SortCode:      sortCode,
		AccountNumber: accountNumber,
	})
	if err != nil {
		return err
	}

	return nil
}
