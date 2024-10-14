package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"strings"
	"time"
)

func parseNotifyApiKey(notifyApiKey string) (string, string) {
	splitKey := strings.Split(notifyApiKey, "-")
	if len(splitKey) != 11 {
		return "", ""
	}
	iss := fmt.Sprintf("%s-%s-%s-%s-%s", splitKey[1], splitKey[2], splitKey[3], splitKey[4], splitKey[5])
	jwtToken := fmt.Sprintf("%s-%s-%s-%s-%s", splitKey[6], splitKey[7], splitKey[8], splitKey[9], splitKey[10])
	return iss, jwtToken
}

func (s *Service) SendEmailToNotify(ctx context.Context, emailAddress string, templateId string) error {
	//notifyUrl := "https://api.notifications.service.gov.uk"
	//emailEndpoint := "v2/notifications/email"

	iss, jwtKey := parseNotifyApiKey(os.Getenv("OPG_NOTIFY_API_KEY"))
	fmt.Println(os.Getenv("OPG_NOTIFY_API_KEY"))
	fmt.Println(iss)
	fmt.Println(jwtKey)

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": iss,
		"iat": time.Now().Unix(),
	})

	signedToken, err := t.SignedString([]byte(jwtKey))
	if err != nil {
		return err
	}

	payload := struct {
		EmailAddress string `json:"email_address"`
		TemplateId   string `json:"template_id"`
	}{
		emailAddress,
		templateId,
	}

	var body bytes.Buffer

	err = json.NewEncoder(&body).Encode(payload)
	if err != nil {
		return err
	}

	//r, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s", notifyUrl, emailEndpoint), &body)
	r, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://google.com", nil)

	if err != nil {
		return err
	}

	r.Header.Add("Content-Type", "application/json")
	r.Header.Add("Authorization", "Bearer "+signedToken)

	resp, err := s.http.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Body)

	return nil
}
