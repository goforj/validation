package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// Item contains an order line's application-owned values.
type Item struct {
	SKU      string
	Quantity int
}

// main collects independent errors using concrete nested paths.
func main() {
	// snippet:nested
	items := []Item{
		{SKU: "book", Quantity: 0},
	}

	err := validation.Validate(
		validation.Group("address",
			validation.String("city", "").Required(),
		),

		validation.Slice("items", items).
			Required().
			Max(100).
			Each(func(item Item) []validation.Field {
				return []validation.Field{
					validation.String("sku", item.SKU).
						Required(),
					validation.Number(
						"quantity",
						item.Quantity,
					).Min(1),
				}
			}),
	)

	failures := err.(*validation.Errors)
	for _, issue := range failures.Issues() {
		fmt.Println(issue.Path, issue.Code)
	}
	// address.city required
	// items.0.quantity min
	// endsnippet:nested

	// Output:
	// address.city required
	// items.0.quantity min
}
