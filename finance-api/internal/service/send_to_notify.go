package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	"os"
	"time"
)

func (s *Service) SendEmailToNotify(ctx context.Context, emailAddress string, templateId string) error {
	notifyUrl := "https://api.notifications.service.gov.uk"
	emailEndpoint := "v2/notifications/email"

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss": os.Getenv("OPG_NOTIFY_API_KEY"),
		"iat": time.Now().Unix(),
	})

	key := os.Getenv("OPG_CORE_JWT_KEY")

	signedToken, err := t.SignedString(key)
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

	r, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s", notifyUrl, emailEndpoint), &body)

	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Authorization", "Bearer "+signedToken)

	resp, err := s.http.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println(resp.Body)

	return nil
}
