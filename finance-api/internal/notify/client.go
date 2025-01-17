package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/apierror"
	"net/http"
	"slices"
	"strings"
	"time"
)

const notifyUrl = "https://api.notifications.service.gov.uk"
const emailEndpoint = "v2/notifications/email"
const ProcessingErrorTemplateId = "872d88b3-076e-495c-bf81-a2be2d3d234c"
const ProcessingFailedTemplateId = "a8f9ab79-1489-4639-9e6c-cad1f079ebcf"
const ProcessingSuccessTemplateId = "8c85cf6c-695f-493a-a25f-77b4fb5f6a8e"

type ProcessingFailedPersonalisation struct {
	FailedLines []string `json:"failed_lines"`
	UploadType  string   `json:"upload_type"`
}

type ProcessingSuccessPersonalisation struct {
	UploadType string `json:"upload_type"`
}

type Payload struct {
	EmailAddress    string      `json:"email_address"`
	TemplateId      string      `json:"template_id"`
	Personalisation interface{} `json:"personalisation"`
}

type Client struct {
	http     *http.Client
	iss      string
	jwtToken string
}

func NewClient(apiKey string) *Client {
	iss, jwtToken := parseNotifyApiKey(apiKey)
	return &Client{
		http:     http.DefaultClient,
		iss:      iss,
		jwtToken: jwtToken,
	}
}

func parseNotifyApiKey(notifyApiKey string) (string, string) {
	splitKey := strings.Split(notifyApiKey, "-")
	if len(splitKey) != 11 {
		return "", ""
	}
	iss := fmt.Sprintf("%s-%s-%s-%s-%s", splitKey[1], splitKey[2], splitKey[3], splitKey[4], splitKey[5])
	jwtToken := fmt.Sprintf("%s-%s-%s-%s-%s", splitKey[6], splitKey[7], splitKey[8], splitKey[9], splitKey[10])
	return iss, jwtToken
}

func (c *Client) Send(ctx context.Context, payload Payload) error {
	signedToken, err := c.createSignedJwtToken()
	if err != nil {
		return err
	}

	var body bytes.Buffer

	err = json.NewEncoder(&body).Encode(payload)
	if err != nil {
		return err
	}

	r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s", notifyUrl, emailEndpoint), &body)

	if err != nil {
		return err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+signedToken)

	resp, err := c.http.Do(r)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusCreated:
		return nil
	case http.StatusUnauthorized:
		return apierror.Unauthorized{}
	case http.StatusBadRequest:
		return apierror.BadRequest{}
	case http.StatusForbidden:
		return apierror.Forbidden{}
	case http.StatusNotFound:
		return apierror.NotFound{}
	case http.StatusInternalServerError:
		return apierror.InternalServer{}
	default:
		return apierror.StatusError{StatusCode: resp.StatusCode}
	}
}

func (c *Client) createSignedJwtToken() (string, error) {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": c.iss,
		"iat": time.Now().Unix(),
	})

	signedToken, err := t.SignedString([]byte(c.jwtToken))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}

func formatFailedLines(failedLines map[int]string) []string {
	var errorMessage string
	var formattedLines []string
	var keys []int
	for i := range failedLines {
		keys = append(keys, i)
	}

	slices.Sort(keys)

	for _, key := range keys {
		failedLine := failedLines[key]
		errorMessage = ""

		switch failedLine {
		case "DATE_PARSE_ERROR":
			errorMessage = "Unable to parse date"
		case "AMOUNT_PARSE_ERROR":
			errorMessage = "Unable to parse amount"
		case "DUPLICATE_PAYMENT":
			errorMessage = "Duplicate payment line"
		case "CLIENT_NOT_FOUND":
			errorMessage = "Could not find a client with this court reference"
		}

		formattedLines = append(formattedLines, fmt.Sprintf("Line %d: %s", key, errorMessage))
	}

	return formattedLines
}
