package main

import (
	"fmt"

	"github.com/goforj/validation"
)

// Registration contains account input already decoded by the application.
type Registration struct {
	Name  string
	Email string
	Age   int
	Terms bool
}

// main validates a complete registration without binding the library to HTTP.
func main() {
	// snippet:quickstart
	req := Registration{
		Name:  "Ada",
		Email: "ada@example.com",
		Age:   36,
		Terms: true,
	}
	err := validation.Validate(
		validation.String("name", req.Name).Required().Min(2).Max(100),
		validation.String("email", req.Email).Required().Email(),
		validation.Number("age", req.Age).Min(18),
		validation.Bool("terms", req.Terms).Accepted(),
	)
	fmt.Println(err)
	// <nil>
	// endsnippet:quickstart

	// Output:
	// <nil>
}
