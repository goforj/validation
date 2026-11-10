package main

import (
	"encoding/json"
	"fmt"

	"github.com/goforj/validation"
)

// main distinguishes patch values from missing properties and explicit null.
func main() {
	// snippet:presence
	var patch struct {
		Seats validation.Input[int]    `json:"seats,omitzero"`
		Name  validation.Input[string] `json:"name,omitzero"`
	}
	fmt.Println(json.Unmarshal([]byte(`{"seats":0,"name":null}`), &patch))
	// <nil>
	fmt.Println(validation.Validate(
		validation.Presence("seats", patch.Seats, func(v int) validation.Field {
			return validation.Number("", v).Min(0)
		}).Required(),
		validation.Presence("name", patch.Name, nonBlank).Nullable(),
	))
	// <nil>
	fmt.Println(patch.Seats.IsPresent(), patch.Name.IsNull())
	// true true
	value, ok := validation.Provided(false).Get()
	fmt.Println(value, ok, validation.Null[string]().IsNull())
	// false true true
	var nickname *string
	fmt.Println(validation.Validate(validation.Optional(nickname, nonBlank)))
	// <nil>
	name := "Ada"
	fmt.Println(validation.Validate(
		validation.Optional(&name, nonBlank),
		validation.Presence("name", validation.Provided(name), nonBlank),
	))
	// <nil>
	// endsnippet:presence
}

// snippet:presence_rule
// nonBlank shares a content rule across optional and presence-aware inputs.
func nonBlank(value string) validation.Field { return validation.String("", value).NotBlank() }

// endsnippet:presence_rule

// Output:
// <nil>
// <nil>
// true true
// false true true
// <nil>
// <nil>
