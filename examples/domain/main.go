package main

import (
	"context"
	"fmt"

	"github.com/goforj/validation"
)

// Delivery is a domain value with an application-specific scheduling policy.
type Delivery struct {
	Days int
}

// main adapts functions into rules without weakening domain type checking.
func main() {
	// snippet:domain
	window := validation.RuleFunc[Delivery](func(value Delivery) error {
		if value.Days > 30 {
			return validation.Invalid("delivery_window", "Choose delivery within 30 days.")
		}
		return nil
	})

	available := validation.ContextRuleFunc[Delivery](
		func(ctx context.Context, _ Delivery) error {
			return ctx.Err()
		},
	)

	err := validation.Validate(
		validation.Value("delivery", Delivery{Days: 7}).
			Using(window).
			Using(available),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:domain
	fmt.Println(window.Validate(Delivery{Days: 40}))
	// Choose delivery within 30 days.

	// Output:
	// <nil>
	// Choose delivery within 30 days.
}
