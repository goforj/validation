package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// Product is a catalog item whose identifier must be unique within an order.
type Product struct{ ID int }

// main demonstrates common checks without defining application-specific validator types.
func main() {
	// snippet:text
	err := validation.Validate(
		validation.String("sku", "BOOK-42").Len(7).Contains("-").NotContains(" ").StartsWith("BOOK").EndsWith("42"),
		validation.String("username", "ada").NotEqual("root").NotIn("admin", "system"),
	)
	fmt.Println(err)
	// <nil>
	// endsnippet:text
	// snippet:numbers
	err = validation.Validate(
		validation.Number("quantity", 12).GreaterThan(0).LessThan(100).MultipleOf(3),
	)
	fmt.Println(err)
	// <nil>
	// endsnippet:numbers
	// snippet:collection_rules
	err = validation.Validate(
		validation.Slice("tags", []string{"go", "books"}).Unique().Contains("go"),
		validation.Slice("products", []Product{{ID: 1}, {ID: 2}}).
			Using(validation.UniqueBy(func(product Product) int { return product.ID })),
		validation.Map("labels", map[string]string{"category": "books"}).
			Required().HasKey("category").OnlyKeys("category").
			Each(func(key, value string) []validation.Field {
				return []validation.Field{validation.String("", value).NotBlank()}
			}),
	)
	fmt.Println(err)
	// <nil>
	// endsnippet:collection_rules

	// Output:
	// <nil>
	// <nil>
	// <nil>
}
