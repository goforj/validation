package main

import (
	"encoding/json"
	"fmt"

	"github.com/goforj/validation"
)

// PaymentRequest models a payment payload decoded by the application.
type PaymentRequest struct {
	Amount   json.Number
	Currency validation.CurrencyCode
}

// main validates the amount and accepted currencies of a payment request.
func main() {
	request := PaymentRequest{
		Amount:   json.Number("1234.50"),
		Currency: validation.USD,
	}

	// snippet:decimal
	err := validation.Validate(
		validation.Decimal("amount", request.Amount).
			Required().
			Min("0.01").
			Max("999999.99").
			MaxScale(2),
		validation.Currency("currency", request.Currency).
			Required().
			USD().
			EUR(),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:decimal
}

// Output:
// <nil>
