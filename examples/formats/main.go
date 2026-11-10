package main

import (
	"fmt"

	"github.com/goforj/validation"
	"github.com/goforj/validation/is"
)

// main shares format rules between standalone and fluent validation.
func main() {
	// snippet:formats
	err := validation.Validate(
		validation.String("request_id", "0190bde7-5ad0-7000-8000-000000000001").Is(is.UUIDv7),
		validation.String("ip", "192.0.2.1").Is(is.IPv4),
		validation.String("metadata", `{"source":"web"}`).Is(is.JSON),
	)
	fmt.Println(err)
	// <nil>
	fmt.Println(is.Email.Validate("ada@example.com"))
	// <nil>
	// endsnippet:formats

	// Output:
	// <nil>
	// <nil>
}
