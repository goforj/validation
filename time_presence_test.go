package validation

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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
