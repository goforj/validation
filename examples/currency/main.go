package main

import (
	"encoding/json"
	"fmt"

	"github.com/goforj/validation"
)

// main decodes a typed currency code and configures one fluent allowlist.
func main() {
	// snippet:currency
	var request struct {
		Currency validation.CurrencyCode `json:"currency"`
	}
	fmt.Println(json.Unmarshal([]byte(`{"currency":"EUR"}`), &request))
	// <nil>
	fmt.Println(validation.Validate(
		validation.Currency("currency", request.Currency).
			Required().USD().EUR().GBP(),
	))
	// <nil>
	// endsnippet:currency
}

// Output:
// <nil>
// <nil>
