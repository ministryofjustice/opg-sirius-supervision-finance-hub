package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-sirius-supervision-finance-hub/shared"
)

type Context struct {
	context.Context
	User *shared.User
}

func NewContext(r *http.Request) Context {
	return Context{
		Context: r.Context(),
	}
}

func (c Context) WithContext(ctx context.Context) Context {
	return Context{
		Context: ctx,
		User:    c.User,
	}
}

type JWT struct {
	Secret string
}

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (j JWT) Verify(requestToken string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(requestToken, &Claims{}, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.Secret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}
