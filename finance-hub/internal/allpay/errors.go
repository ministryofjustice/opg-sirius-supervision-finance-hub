package allpay

import (
	"errors"
	"fmt"
	"strings"
)

var ErrorAPI = errors.New("AllPay API error")

type ErrorValidation struct {
	Messages []string `json:"messages"`
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("validation: %s", strings.Join(e.Messages, ", "))
}
