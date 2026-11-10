package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// main applies format checks directly to fluent string chains.
func main() {
	// snippet:formats
	err := validation.Validate(
		validation.String("email", "ada@example.com").IsEmail(),
		validation.String("site", "https://goforj.dev").IsURL(),

		validation.String("id", "550e8400-e29b-41d4-a716-446655440000").
			IsUUID(),
		validation.String("trace_id", "550e8400-e29b-41d4-a716-446655440000").
			IsUUIDv4(),
		validation.String("request_id", "0190bde7-5ad0-7000-8000-000000000001").
			IsUUIDv7(),

		validation.String("host", "::1").IsIP(),
		validation.String("ip", "192.0.2.1").IsIPv4(),
		validation.String("ipv6", "2001:db8::1").IsIPv6(),

		validation.String("metadata", `{"source":"web"}`).IsJSON(),
		validation.String("cursor", "YWJj").IsBase64(),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:formats

	// Output:
	// <nil>
}
