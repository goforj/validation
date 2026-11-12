package main

import (
	"errors"
	"fmt"

	"github.com/goforj/validation"
)

// Registration contains account input already decoded by the application.
type Registration struct {
	Name                 string
	Email                string
	Password             string
	PasswordConfirmation string
	AccountType          string
	Company              string
	Age                  int
	Terms                bool
}

// main demonstrates grouped registration rules and useful errors for invalid input.
func main() {
	// snippet:quickstart
	err := validation.Validate(
		validation.String("name", "Ada").Required().Min(2),
		validation.String("email", "ada@example.com").Required().IsEmail(),
		validation.Number("age", 36).Min(18),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:quickstart

	req := Registration{
		Name:                 "Ada",
		Email:                "not-an-email",
		Password:             "long-enough-password",
		PasswordConfirmation: "a-different-password",
		AccountType:          "business",
		Company:              "",
		Age:                  36,
		Terms:                false,
	}

	// snippet:registration
	registrationErr := validation.Validate(
		// Identity and contact details.
		validation.String("name", req.Name).Required().Min(2).Max(100),
		validation.String("email", req.Email).Required().IsEmail().
			Message("Enter a valid email address."),

		// Password length and confirmation.
		validation.String("password", req.Password).Required().Min(12),
		validation.String("password_confirmation", req.PasswordConfirmation).
			Required().Equal(req.Password),

		// Business accounts must provide a company name.
		validation.String("account_type", req.AccountType).In("personal", "business"),
		validation.String("company", req.Company).
			RequiredIf(req.AccountType == "business"),

		// Eligibility and consent.
		validation.Number("age", req.Age).Min(18),
		validation.Bool("terms", req.Terms).Accepted(),
	)
	// endsnippet:registration

	// snippet:quickstart_errors
	var failures *validation.Errors
	if errors.As(registrationErr, &failures) {
		fmt.Println(failures.First("email"))
		// Enter a valid email address.

		for _, issue := range failures.Issues() {
			fmt.Println(issue.Path, issue.Code)
		}
		// email email
		// password_confirmation equal
		// company required
		// terms accepted
	}
	// endsnippet:quickstart_errors
}

// Output:
// <nil>
// Enter a valid email address.
// email email
// password_confirmation equal
// company required
// terms accepted
