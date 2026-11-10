package main

import (
	"encoding/json"
	"fmt"

	"github.com/goforj/validation"
)

// snippet:decimal_types
// Currency is the application's currency-code type; validation maintains no currency registry.
type Currency string

const (
	// USD is an application-supported currency code.
	USD Currency = "USD"
	// EUR is an application-supported currency code.
	EUR Currency = "EUR"
)

// Invoice preserves decimal text and a typed currency code after application decoding.
type Invoice struct {
	Amount   json.Number
	Currency Currency
}

// endsnippet:decimal_types

// main applies exact decimal constraints and an application-owned typed currency allowlist.
func main() {
	// snippet:decimal
	invoice := Invoice{Amount: json.Number("1234.50"), Currency: USD}
	fmt.Println(validation.Validate(
		validation.Decimal("amount", invoice.Amount).
			Required().Min("0.01").Max("999999.99").MaxScale(2),
		validation.String("currency", invoice.Currency).In(USD, EUR),
	))
	// <nil>
	fmt.Println(validation.Validate(
		validation.Decimal("amount", "9007199254740993.01").
			GreaterThan("9007199254740993").LessThan("9007199254740994").
			Equal("9007199254740993.010"),
	))
	// <nil>
	// endsnippet:decimal
}

// Output:
// <nil>
// <nil>
