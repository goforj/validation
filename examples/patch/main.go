package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// ProfilePatch uses nil to represent an omitted email address.
type ProfilePatch struct {
	Email *string
}

// Validate checks only fields supplied by the application decoder.
func (req ProfilePatch) Validate() error {
	// snippet:patch
	return validation.Validate(
		validation.When(req.Email != nil, func() validation.Field {
			return validation.String("email", *req.Email).Required().IsEmail()
		}),
	)
	// endsnippet:patch
}

// main demonstrates both omitted and supplied values using the same validation method.
func main() {
	fmt.Println((ProfilePatch{}).Validate())
	// <nil>
	email := "ada@example.com"
	fmt.Println((ProfilePatch{Email: &email}).Validate())
	// <nil>

	// Output:
	// <nil>
	// <nil>
}
