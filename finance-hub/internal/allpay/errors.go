package allpay

import (
	"fmt"
	"strings"
)

type ErrorAPI struct{}

func (e ErrorAPI) Error() string {
	return fmt.Sprintf("Direct debit cannot be setup due to an unexpected response from AllPay.")
}

type ErrorValidation struct {
	Messages []string `json:"messages"`
}

func (e ErrorValidation) Error() string {
	return fmt.Sprintf("validation: %s", strings.Join(e.Messages, ", "))
}

type ErrorModulusCheckFailed struct{}

func (e ErrorModulusCheckFailed) Error() string {
	return fmt.Sprintf("Modulus check on account and sort code failed. Please check details are correct.")
}
