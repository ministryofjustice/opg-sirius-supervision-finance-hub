package middleware

import (
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
)

// JwtCookieVerifier verifies the JWT stored as a token in the session cookie issued by Membrane
type JwtCookieVerifier struct {
	Secret string
}

func (j *JwtCookieVerifier) Verify() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, _ := r.Cookie("OPG-TOKEN")

			token, verifyErr := VerifyToken(cookie, j.Secret)

			if verifyErr != nil {
				log.Println("Error in token verification :", verifyErr.Error())
				w.WriteHeader(http.StatusUnauthorized)
			} else {
				claims := token.Claims.(jwt.MapClaims)
				email := claims["session-data"].(string)
				log.Println(email)
				//ctx := context.WithValue(r.Context(), HashedEmail{}, hashedEmail)
				next.ServeHTTP(w, r)
			}
		})
	}
}

func VerifyToken(cookie *http.Cookie, secret string) (*jwt.Token, error) {
	if cookie == nil {
		return nil, errors.New("missing authentication token")
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (i interface{}, err error) {
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
