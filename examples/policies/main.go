package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/goforj/validation"
)

// main demonstrates reusable rule composition and configurable evaluation policies.
func main() {
	composition()
	policies()
}

// composition builds reusable alternatives before attaching them to application values.
func composition() {
	// snippet:composition_rules
	startsA := validation.RuleFunc[string](
		func(value string) error {
			if strings.HasPrefix(value, "A") {
				return nil
			}
			return validation.Invalid("starts_a", "Must start with A")
		},
	)

	endsZ := validation.ContextRuleFunc[string](
		func(_ context.Context, value string) error {
			if strings.HasSuffix(value, "Z") {
				return nil
			}
			return validation.Invalid("ends_z", "Must end with Z")
		},
	)

	allowedCode := validation.All(
		validation.Any[string](startsA, endsZ),
		validation.Not[string](endsZ),
	)
	// endsnippet:composition_rules

	// snippet:composition
	err := validation.Validate(
		validation.String("code", "ABC").Using(allowedCode),
	)

	fmt.Println(err)
	// <nil>

	err = validation.Validate(
		validation.String("code", "XYZ").
			Using(validation.Any[string](startsA, endsZ)),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:composition
}

// policies changes error collection and presentation without rebuilding field rules.
func policies() {
	// snippet:policies
	validator := validation.New().
		CollectAll().
		FormatMessages(func(issue validation.Issue) string {
			return issue.Path + ": " + issue.Code
		})

	err := validator.Validate(
		validation.String("name", "a").Min(3).Contains("z"),
	)

	var failures *validation.Errors
	if errors.As(err, &failures) {
		fmt.Println(failures.Messages()["name"])
		// [name: min name: contains]
	}

	first := validator.StopOnFirstError()

	err = first.ValidateContext(
		context.Background(),
		validation.String("name", "").Required(),
		validation.Number("age", 0).Min(18),
	)

	fmt.Println(err)
	// name: required
	// endsnippet:policies
}

// Output:
// <nil>
// <nil>
// [name: min name: contains]
// name: required
