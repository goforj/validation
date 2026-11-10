package main

import (
	"fmt"
	"strings"

	"github.com/goforj/validation"
)

// snippet:custom_type
// AllowedDomain requires the configured company domain.
type AllowedDomain struct {
	Domain string
}

// Validate rejects addresses outside the company.
func (r AllowedDomain) Validate(value string) error {
	_, domain, ok := strings.Cut(value, "@")
	if !ok || !strings.EqualFold(domain, r.Domain) {
		return validation.Invalid("email_domain", "Use your company email address.")
	}
	return nil
}

// endsnippet:custom_type

// main attaches an application-defined rule without registration or reflection.
func main() {
	// snippet:custom
	err := validation.Validate(
		validation.String("email", "ada@other.example").
			Required().
			IsEmail().
			Using(AllowedDomain{Domain: "goforj.dev"}),
	)

	fmt.Println(err)
	// Use your company email address.
	// endsnippet:custom
	// snippet:invalid
	fmt.Println(validation.Invalid("email_domain", "Use your company email address."))
	// Use your company email address.
	// endsnippet:invalid
	fmt.Println((AllowedDomain{Domain: "goforj.dev"}).Validate("ada@goforj.dev"))
	// <nil>

	// Output:
	// Use your company email address.
	// Use your company email address.
	// <nil>
}
