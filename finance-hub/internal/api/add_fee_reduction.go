package api

import (
	"bytes"
	"encoding/json"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
)

func (c *ApiClient) AddFeeReduction(ctx Context, financeClientId int, feeType string, startYear string, lengthOfAward string, dateReceived string, feeReductionNotes string) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(shared.AddFeeReduction{
		FinanceClientId:   financeClientId,
		FeeType:           feeType,
		StartYear:         startYear,
		LengthOfAward:     lengthOfAward,
		DateReceive:       dateReceived,
		FeeReductionNotes: feeReductionNotes,
	})
	if err != nil {
		return err
	}

	req, err := c.newBackendRequest(ctx, http.MethodPost, "/fee-reductions", &body)

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
