package main

import (
	"errors"
	"fmt"

	"github.com/goforj/validation"
)

// main exposes safe structured errors for an application's form or API layer.
func main() {
	// snippet:errors
	err := validation.Validate(
		validation.String("email", "bad").Label("email address").Email().
			Message("Enter your work email address."),
	)
	var failures *validation.Errors
	if errors.As(err, &failures) {
		fmt.Println(failures.First("email"))
		// Enter your work email address.
		fmt.Println(failures.Messages()["email"])
		// [Enter your work email address.]
		fmt.Println(failures.Issues()[0].Code)
		// email
	}
	// endsnippet:errors

	// Output:
	// Enter your work email address.
	// [Enter your work email address.]
	// email
}
