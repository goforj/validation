package validation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

// TestTimeRules checks inclusive and exclusive instant bounds and date parsing policies.
func TestTimeRules(t *testing.T) {
	now := time.Date(2026, 11, 10, 0, 0, 0, 0, time.UTC)
	good := Time("at", now).Label("start").Required().RequiredIf(false).Before(now.Add(time.Second)).After(now.Add(-time.Second)).Min(now).Max(now).Equal(now.In(time.FixedZone("local", 3600))).Using(RuleFunc[time.Time](func(time.Time) error { return nil })).Check(func(context.Context, time.Time) error { return nil }).Message("custom")
	if err := Validate(good, String("date", "2024-02-29").IsDate(time.DateOnly), String("local", "2026-11-10 12:30").IsDateIn("2006-01-02 15:04", time.FixedZone("local", 3600))); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		f    Field
		code string
	}{
		{Time("at", time.Time{}).Required(), "required"}, {Time("at", now).Before(now), "before"}, {Time("at", now).After(now), "after"},
		{Time("at", now).Min(now.Add(time.Second)), "min"}, {Time("at", now).Max(now.Add(-time.Second)), "max"}, {Time("at", now).Equal(now.Add(time.Second)), "equal"}, {String("d", "2023-02-29").IsDate(time.DateOnly), "date"},
	} {
		failureAt(t, tc.f, tc.code)
	}
	for _, f := range []Field{String("d", "").Required().IsDate(""), String("d", "").IsDateIn(time.DateOnly, nil), Time("at", now).Using(nil), Time("at", now).Message("bad")} {
		var config *ConfigurationError
		if err := Validate(f); !errors.As(err, &config) {
			t.Fatal(err)
		}
	}
}

// TestInputJSON preserves omitted, null, and zero states and failed-decode atomicity.
func TestInputJSON(t *testing.T) {
	type request struct {
		Count  Input[int]    `json:"count,omitzero"`
		Active Input[bool]   `json:"active,omitzero"`
		Name   Input[string] `json:"name,omitzero"`
	}
	var req request
	if err := json.Unmarshal([]byte(`{"count":0,"active":false,"name":null}`), &req); err != nil {
		t.Fatal(err)
	}
	if v, ok := req.Count.Get(); !ok || v != 0 || !req.Count.IsPresent() || req.Count.IsNull() {
		t.Fatal(req.Count)
	}
	if v, ok := req.Active.Get(); !ok || v {
		t.Fatal(req.Active)
	}
	if !req.Name.IsNull() || !req.Name.IsPresent() {
		t.Fatal(req.Name)
	}
	if _, ok := req.Name.Get(); ok {
		t.Fatal("null has value")
	}
	got, err := json.Marshal(req)
	if err != nil || string(got) != `{"count":0,"active":false,"name":null}` {
		t.Fatalf("%s: %v", got, err)
	}
	got, err = json.Marshal(request{})
	if err != nil || string(got) != "{}" {
		t.Fatalf("%s: %v", got, err)
	}
	var absent Input[int]
	if absent.IsPresent() || !absent.IsZero() {
		t.Fatal(absent)
	}
	got, err = json.Marshal(absent)
	if err != nil || string(got) != "null" {
		t.Fatal(string(got), err)
	}
	value := Provided(7)
	if err := value.UnmarshalJSON([]byte(`"wrong"`)); err == nil {
		t.Fatal("invalid input accepted")
	}
	if v, ok := value.Get(); !ok || v != 7 {
		t.Fatal(value)
	}
	if err := value.UnmarshalJSON([]byte(" \nnull\t")); err != nil || !value.IsNull() {
		t.Fatal(value, err)
	}
	if _, err := json.Marshal(Provided(make(chan int))); err == nil {
		t.Fatal("unsupported value encoded")
	}
}

// TestPresenceAndOptional checks lazy callbacks, zero values, nesting, and policy precedence.
func TestPresenceAndOptional(t *testing.T) {
	calls := 0
	build := func(v int) Field { calls++; return Number("", v).Min(0) }
	var missing Input[int]
	zero := 0
	if err := Validate(Presence("count", missing, build), Presence("count", Null[int](), build).Nullable(), Presence("count", Provided(0), build).Required(), Presence("count", missing, build).RequiredIf(false), Optional[int](nil, build), Optional(&zero, build)); err != nil || calls != 2 {
		t.Fatal(err, calls)
	}
	failureAt(t, Presence("count", missing, build).Required(), "required")
	failureAt(t, Presence("count", Null[int](), build), "null")
	failureAt(t, Presence("count", Null[int](), build).Nullable().Required(), "required")
	var failures *Errors
	err := Validate(Group("request", Presence("count", missing, build).Label("Count").Required().Message("Supply count")))
	if !errors.As(err, &failures) || !reflect.DeepEqual(failures.Issues(), []Issue{{Path: "request.count", Code: "required", Message: "Supply count"}}) {
		t.Fatal(err)
	}
	failureAt(t, Presence("count", Provided(-1), build), "min")
	for _, f := range []Field{Presence[int]("count", missing, nil), Presence("count", missing, build).Message("bad"), Optional[int](nil, nil), Optional(&zero, func(int) Field { return nil }), Presence("count", Provided(0), func(int) Field { return nil })} {
		var config *ConfigurationError
		if err := Validate(f); !errors.As(err, &config) {
			t.Fatal(err)
		}
	}
	op := errors.New("unavailable")
	if err := Validate(Presence("count", Provided(1), func(v int) Field { return Value("", v).Check(func(context.Context, int) error { return op }) })); !errors.Is(err, op) {
		t.Fatal(err)
	}
}

// TestInputRejectsNonJSONWhitespace preserves the previous state when null has invalid surrounding bytes.
func TestInputRejectsNonJSONWhitespace(t *testing.T) {
	for _, data := range []string{"\vnull", "null\f", "\u00a0null", "null\u0085"} {
		input := Provided(7)
		if err := input.UnmarshalJSON([]byte(data)); err == nil {
			t.Errorf("accepted malformed JSON: %q", data)
		}
		if value, ok := input.Get(); !ok || value != 7 {
			t.Errorf("failed decode changed previous value: %q", data)
		}
	}
}

// FuzzInputJSON checks parsing validity and failed-decode atomicity against the JSON scanner.
func FuzzInputJSON(f *testing.F) {
	for _, value := range []string{"null", " \t\r\nnull\n", "\vnull", "null\u00a0", "0", "false", `{"name":"Ada"}`, `"text"`, "[1,null]", "1e10000", "[", ""} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, data string) {
		previous := json.RawMessage(`"previous"`)
		input := Provided(previous)
		err := input.UnmarshalJSON([]byte(data))
		if (err == nil) != json.Valid([]byte(data)) {
			t.Fatalf("JSON validity differs for %q: %v", data, err)
		}
		if err != nil {
			if value, ok := input.Get(); !ok || string(value) != string(previous) {
				t.Fatal("failed decode replaced the previous value")
			}
			return
		}
		if !input.IsPresent() {
			t.Fatal("successful decode lost presence")
		}
		encoded, err := input.MarshalJSON()
		if err != nil || !json.Valid(encoded) {
			t.Fatalf("successful decode did not marshal valid JSON: %s, %v", encoded, err)
		}
	})
}

// TestInputDecoderBoundaries documents typed precision and the scope of enclosing decoder options.
func TestInputDecoderBoundaries(t *testing.T) {
	var request struct {
		Payload Input[struct {
			Count uint64 `json:"count"`
		}] `json:"payload"`
		Plain   any        `json:"plain"`
		Wrapped Input[any] `json:"wrapped"`
	}
	decoder := json.NewDecoder(strings.NewReader(`{"payload":{"count":18446744073709551615,"unknown":true},"plain":7,"wrapped":7}`))
	decoder.DisallowUnknownFields()
	decoder.UseNumber()
	if err := decoder.Decode(&request); err != nil {
		t.Fatal(err)
	}
	payload, ok := request.Payload.Get()
	if !ok || payload.Count != ^uint64(0) {
		t.Fatal("typed integer precision changed")
	}
	if _, ok := request.Plain.(json.Number); !ok {
		t.Fatal("enclosing decoder lost UseNumber")
	}
	wrapped, ok := request.Wrapped.Get()
	if _, number := wrapped.(float64); !ok || !number {
		t.Fatal("Input no longer follows json.Unmarshal defaults")
	}
}

// TestTimeEqualityRedactsExpectedInstant keeps equality targets private like string and numeric equality.
func TestTimeEqualityRedactsExpectedInstant(t *testing.T) {
	expected := time.Date(1980, 1, 2, 3, 4, 5, 0, time.UTC)
	var failures *Errors
	if err := Validate(Time("confirmed_at", expected.Add(time.Second)).Equal(expected)); !errors.As(err, &failures) {
		t.Fatal(err)
	}
	issue := failures.Issues()[0]
	if issue.Code != "equal" || issue.Parameters != nil {
		t.Fatalf("equality retained expected instant: %+v", issue)
	}
}
