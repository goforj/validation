package main

import (
	"fmt"
	"regexp"

	"github.com/goforj/validation"
)

// main demonstrates explicit value constraints without automatic coercion or empty-value skipping.
func main() {
	// snippet:constraints
	skuPattern := regexp.MustCompile(`^[A-Z]+-\d+$`)

	err := validation.Validate(
		validation.String("name", "Ada").Required().NotBlank().Min(2).Max(100),

		validation.String("company", "").RequiredIf(false).Max(200),
		validation.String("role", "member").In("member", "admin"),

		validation.String("confirmation", "correct horse").Equal("correct horse"),
		validation.String("sku", "BOOK-42").Match(skuPattern),

		validation.Number("age", 36).Min(18).Max(120),
		validation.Bool("terms", true).Accepted(),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:constraints

	// Output:
	// <nil>
}
