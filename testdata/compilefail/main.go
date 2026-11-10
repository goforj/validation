package main

import "github.com/goforj/validation"

// main contains intentional type errors checked by TestInvalidAPIsDoNotCompile.
func main() {
	validation.Number("age", 18).Email()
	validation.Number("age", 18).Min(1.5)
	validation.Bool("active", false).Required()
	validation.String("email", "value").Using(validation.RuleFunc[int](func(int) error { return nil }))
}
