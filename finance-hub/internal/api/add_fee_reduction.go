package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"time"
)

func (c *ApiClient) AddFeeReduction(ctx Context, clientId int, feeType string, startYear string, lengthOfAward string, dateReceived string, feeReductionNotes string) error {
	var body bytes.Buffer

	dateReceivedTransformed, _ := time.Parse("2006-01-02", dateReceived)
	err := json.NewEncoder(&body).Encode(shared.AddFeeReduction{
		ClientId:          clientId,
		FeeType:           feeType,
		StartYear:         startYear,
		LengthOfAward:     lengthOfAward,
		DateReceive:       shared.Date{Time: dateReceivedTransformed},
		FeeReductionNotes: feeReductionNotes,
	})
	if err != nil {
		return err
	}

	url := fmt.Sprintf("/clients/%d/fee-reductions", clientId)
	req, err := c.newBackendRequest(ctx, http.MethodPost, url, &body)

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
		var v ValidationError
		if err := json.NewDecoder(resp.Body).Decode(&v); err == nil && len(v.Errors) > 0 {
			return ValidationError{Errors: v.Errors}
		}
		return newStatusError(resp)
	}

	return nil
}
