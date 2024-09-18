package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/opg-sirius-finance-hub/apierror"
	"github.com/opg-sirius-finance-hub/shared"
	"net/http"
	"strconv"
	"time"
)

func (c *ApiClient) AddFeeReduction(ctx Context, clientId int, feeType string, startYear string, lengthOfAward string, dateReceived string, notes string) error {
	var body bytes.Buffer
	var dateReceivedTransformed *shared.Date

	if dateReceived != "" {
		dateReceivedToTime, _ := time.Parse("2006-01-02", dateReceived)
		dateReceivedTransformed = &shared.Date{Time: dateReceivedToTime}
	}

	lengthOfAwardTransformed, _ := strconv.Atoi(lengthOfAward)
	err := json.NewEncoder(&body).Encode(shared.AddFeeReduction{
		FeeType:       shared.ParseFeeReductionType(feeType),
		StartYear:     startYear,
		LengthOfAward: lengthOfAwardTransformed,
		DateReceived:  dateReceivedTransformed,
		Notes:         notes,
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

	if resp.StatusCode == http.StatusBadRequest {
		return apierror.ValidationError{Errors: apierror.ValidationErrors{"Overlap": {"start-or-end-date": ""}}}
	}

	return newStatusError(resp)
}
