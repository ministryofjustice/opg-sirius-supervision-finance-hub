package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

type JwtConfig struct {
	Enabled bool
	Secret  string
	Expiry  int
}

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (j JwtConfig) Verify(requestToken string) (*jwt.Token, error) {
	if !j.Enabled {
		return &jwt.Token{}, nil
	}
	token, err := jwt.ParseWithClaims(requestToken, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})

	if err != nil {
		return nil, errors.New(err.Error())
	}

	return token, nil
}

func (j JwtConfig) CreateToken(clientId int) (accessToken string, err error) {
	if !j.Enabled {
		return "", nil
	}
	exp := time.Now().Add(time.Second * time.Duration(j.Expiry))
	claims := &Claims{
		Roles: []string{"urn:opg:sirius:private-finance-manager"},
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        strconv.Itoa(clientId),
			Issuer:    "urn:opg:payments-hub",
			Audience:  jwt.ClaimStrings{"urn:opg:payments-api", "urn:opg:sirius"},
			Subject:   "urn:opg:sirius:users:" + strconv.Itoa(clientId),
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
