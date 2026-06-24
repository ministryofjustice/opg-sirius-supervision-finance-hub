package allpay

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

type UpdateClientDetailsInput struct {
	ClientDetails
	NewSurname string  `json:"LastName"`
	Address    Address `json:"Address"`
}

type updateClientDetailsRequest struct {
	Address struct {
		LastName      string `json:"LastName"`
		TitleInitials string `json:"TitleInitials"` // unused but required for API
		Address
	} `json:"Address"`
}

func (c *Client) UpdateClientDetails(ctx context.Context, data *UpdateClientDetailsInput) error {
	logger := c.logger(ctx)

	body := updateClientDetailsRequest{}
	body.Address.LastName = trimChars(data.NewSurname, 19)
	body.Address.Address = Address{
		Line1:    trimChars(data.Address.Line1, 40),
		Town:     trimChars(data.Address.Town, 40),
		PostCode: trimChars(data.Address.PostCode, 10),
	}

	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(body)
	if err != nil {
		logger.Error("unable to parse update client details request", "error", err)
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPut,
		fmt.Sprintf("/Customers/%s/%s/%s",
			c.schemeCode,
			base64.StdEncoding.EncodeToString([]byte(data.ClientReference)),
			base64.StdEncoding.EncodeToString([]byte(trimChars(data.Surname, 19))),
		), &buf)

	if err != nil {
		logger.Error("unable to build update client details request", "error", err)
		return ErrorAPI{}
	}

	resp, err := c.http.Do(req)
	if err != nil {
		logger.Error("unable to send update client details request", "error", err)
		return ErrorAPI{}
	}

	defer unchecked(resp.Body.Close)

	if resp.StatusCode == http.StatusUnprocessableEntity {
		var ve ErrorValidation

		err = json.NewDecoder(resp.Body).Decode(&ve)
		if err != nil {
			logger.Error("unable to parse update client details validation response", "error", err)
			return ErrorAPI{}
		}

		logger.Error("update client details request returned validation errors", "errors", ve)
		return ve
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error("update client details request returned unexpected status code", "status", resp.Status)
		return ErrorAPI{}
	}

	return nil
}
