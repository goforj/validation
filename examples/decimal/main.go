package main

import (
	"encoding/json"
	"fmt"

	"github.com/goforj/validation"
)

// snippet:decimal_types
// Invoice preserves decimal text and a typed currency code after application decoding.
type Invoice struct {
	Amount   json.Number
	Currency validation.CurrencyCode
}

// endsnippet:decimal_types

// main applies exact decimal constraints and a fluent typed currency allowlist.
func main() {
	// snippet:decimal
	invoice := Invoice{Amount: json.Number("1234.50"), Currency: validation.USD}
	fmt.Println(validation.Validate(
		validation.Decimal("amount", invoice.Amount).
			Required().Min("0.01").Max("999999.99").MaxScale(2),
		validation.Currency("currency", invoice.Currency).USD().EUR(),
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
