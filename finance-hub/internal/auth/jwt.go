package auth

import (
	"context"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-go-common/telemetry"
)

const expiry = 5

type JWT struct {
	Secret string `json:"-"`
}

type Claims struct {
	Roles []string `json:"roles"`
	jwt.RegisteredClaims
}

func (j *JWT) CreateJWT(ctx context.Context) string {
	user := ctx.(Context).User

	exp := time.Now().Add(time.Second * time.Duration(expiry))
	claims := &Claims{
		Roles: user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        strconv.Itoa(int(user.ID)),
			Issuer:    "urn:opg:payments-hub",
			Audience:  jwt.ClaimStrings{"urn:opg:payments-api"},
			Subject:   "urn:opg:sirius:users:" + strconv.Itoa(int(user.ID)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.Secret))
	if err != nil {
		telemetry.LoggerFromContext(ctx).Error("Error creating JWT", "error", err)
		return ""
	}
	return t
}
