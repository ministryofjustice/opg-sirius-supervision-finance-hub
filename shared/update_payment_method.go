package shared

type UpdatePaymentMethod struct {
	PaymentMethod PaymentMethod `json:"paymentMethod" validate:"valid-enum"` // DEMANDED, DIRECT DEBIT
}
