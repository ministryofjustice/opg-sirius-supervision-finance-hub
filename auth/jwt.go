package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/opg-sirius-finance-hub/shared"
	"strconv"
	"time"
)

type JwtConfig struct {
	Secret string
	Expiry int
}

type Claims struct {
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (j JwtConfig) createToken(user *shared.Assignee) (accessToken string, err error) {
	exp := time.Now().Add(time.Second * time.Duration(j.Expiry))
	claims := &Claims{
		ID:    strconv.Itoa(user.Id),
		Roles: user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "urn:opg:payments-hub",
			Audience:  jwt.ClaimStrings{"urn:opg:payments-api", "urn:opg:sirius"},
			Subject:   "urn:opg:sirius:users:" + strconv.Itoa(user.Id),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		return "", err
	}
	return t, err
}

func (j JwtConfig) IsAuthorized(requestToken string) (bool, error) {
	_, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func ExtractIDFromToken(requestToken string, secret string) (string, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)

	if !ok && !token.Valid {
		return "", fmt.Errorf("Invalid Token")
	}

	return claims["id"].(string), nil
}
