package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/goforj/validation"
	"strings"
)

// main composes reusable rules and configures error presentation and collection.
func main() {
	// snippet:composition
	startsA := validation.RuleFunc[string](func(v string) error {
		if strings.HasPrefix(v, "A") {
			return nil
		}
		return validation.Invalid("starts_a", "Must start with A")
	})
	endsZ := validation.ContextRuleFunc[string](func(_ context.Context, v string) error {
		if strings.HasSuffix(v, "Z") {
			return nil
		}
		return validation.Invalid("ends_z", "Must end with Z")
	})
	fmt.Println(validation.Validate(validation.String("code", "ABC").Using(validation.All(validation.Any[string](startsA, endsZ), validation.Not[string](endsZ)))))
	// <nil>
	fmt.Println(validation.Validate(validation.String("code", "XYZ").Using(validation.Any[string](startsA, endsZ))))
	// <nil>
	// endsnippet:composition
	// snippet:policies
	validator := validation.New().CollectAll().FormatMessages(func(issue validation.Issue) string { return issue.Path + ": " + issue.Code })
	var failures *validation.Errors
	if err := validator.Validate(validation.String("name", "a").Min(3).Contains("z")); errors.As(err, &failures) {
		fmt.Println(failures.Messages()["name"])
		// [name: min name: contains]
	}
	first := validator.StopOnFirstError()
	fmt.Println(first.ValidateContext(context.Background(), validation.String("name", "").Required(), validation.Number("age", 0).Min(18)))
	// name: required
	// endsnippet:policies
}

// Output:
// <nil>
// <nil>
// [name: min name: contains]
// name: required
