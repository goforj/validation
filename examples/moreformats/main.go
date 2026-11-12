package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// main checks common identifiers, network values, and explicit character policies.
func main() {
	// snippet:moreformats
	err := validation.Validate(
		// Sortable identifiers.
		validation.String("id", "01ARZ3NDEKTSV4RRFFQ69G5FAV").IsULID(),

		// Network names and addresses.
		validation.String("host", "api.example.com").IsHostname(),
		validation.String("network", "192.0.2.0/24").IsCIDR(),
		validation.String("device", "00:11:22:33:44:55").IsMAC(),

		// Phone numbers and URL-safe encoding.
		validation.String("phone", "+14155552671").IsE164(),
		validation.String("token", "__8").IsBase64URL(),

		// Character restrictions.
		validation.String("name", "Ada").IsAlpha(),
		validation.String("sku", "A123").IsAlphanumeric(),
		validation.String("pin", "012345").IsDigits(),
		validation.String("key", "header_name").IsASCII(),

		// Unicode casing and text.
		validation.String("slug", "café").IsLowercase(),
		validation.String("code", "ÉCOLE").IsUppercase(),
		validation.String("title", "こんにちは").IsUTF8(),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:moreformats
}

// Output:
// <nil>
