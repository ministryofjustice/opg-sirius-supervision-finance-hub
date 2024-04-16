package auth

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strconv"
	"time"
)

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func Verify(requestToken string, secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(requestToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, errors.New(err.Error())
	}

	return token, nil
}

func CreateToken(clientId int, secret string, expiry int) (accessToken string, err error) {
	exp := time.Now().Add(time.Second * time.Duration(expiry))
	claims := &Claims{
		//Roles: user.Roles,
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
	t, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return t, err
}
