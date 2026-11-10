package main

import (
	"encoding/json"
	"fmt"

	"github.com/goforj/validation"
)

// main distinguishes patch presence, explicit values, and optional pointers.
func main() {
	presence()
	providedValues()
	optionalValues()
}

// presence preserves zero and null states while validating a decoded patch.
func presence() {
	// snippet:presence
	var patch struct {
		Seats validation.Input[int]    `json:"seats"`
		Name  validation.Input[string] `json:"name"`
	}

	payload := []byte(`{"seats":0,"name":null}`)
	decodeErr := json.Unmarshal(payload, &patch)
	fmt.Println(decodeErr)
	// <nil>

	err := validation.Validate(
		validation.Presence(
			"seats",
			patch.Seats,
			func(value int) validation.Field {
				return validation.Number("", value).Min(0)
			},
		).Required(),

		validation.Presence(
			"name",
			patch.Name,
			nonBlank,
		).Nullable(),
	)

	fmt.Println(err)
	// <nil>

	fmt.Println(patch.Seats.IsPresent())
	// true

	fmt.Println(patch.Name.IsNull())
	// true
	// endsnippet:presence
}

// providedValues demonstrates that false is present and null is a separate state.
func providedValues() {
	// snippet:presence_values
	value, present := validation.Provided(false).Get()
	fmt.Println(value, present)
	// false true

	fmt.Println(validation.Null[string]().IsNull())
	// true
	// endsnippet:presence_values
}

// optionalValues reuses one content rule for absent and supplied values.
func optionalValues() {
	// snippet:optional
	var nickname *string

	err := validation.Validate(
		validation.Optional(nickname, nonBlank),
	)

	fmt.Println(err)
	// <nil>

	name := "Ada"

	err = validation.Validate(
		validation.Optional(&name, nonBlank),
	)

	fmt.Println(err)
	// <nil>
	// endsnippet:optional
}

// snippet:presence_rule
// nonBlank reuses a nonblank string rule.
func nonBlank(value string) validation.Field {
	return validation.String("", value).NotBlank()
}

// endsnippet:presence_rule

// Output:
// <nil>
// <nil>
// true
// true
// false true
// true
// <nil>
// <nil>
