package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// main validates the accepted currencies of a decoded request.
func main() {
	request := struct {
		Currency validation.CurrencyCode
	}{Currency: validation.EUR}

	// snippet:currency
	err := validation.Validate(
		validation.Currency("currency", request.Currency).
			Required().
			USD().
			EUR().
			GBP(),
	)
	fmt.Println(err)
	// <nil>
	// endsnippet:currency
}

// Output:
// <nil>
