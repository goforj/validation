package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/goforj/validation"
)

// main keeps an unavailable application dependency separate from invalid input.
func main() {
	// snippet:context
	unavailable := errors.New("directory unavailable")

	err := validation.ValidateContext(
		context.Background(),
		validation.String("email", "ada@example.com").
			IsEmail().
			Check(
				func(ctx context.Context, v string) error {
					// A client would use ctx and v.
					return unavailable
				},
			),
	)

	fmt.Println(errors.Is(err, unavailable))
	// true
	// endsnippet:context

	// Output:
	// true
}
