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
	err := validation.ValidateContext(context.Background(),
		validation.String("email", "ada@example.com").Email().
			Check(func(ctx context.Context, email string) error {
				// A real directory client would use ctx and email here.
				return unavailable
			}),
	)
	fmt.Println(errors.Is(err, unavailable))
	// true
	// endsnippet:context

	// Output:
	// true
}
