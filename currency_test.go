package validation

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// TestCurrencySelectors checks every generated selector against the independently pinned identifiers.
func TestCurrencySelectors(t *testing.T) {
	data, err := os.ReadFile("testdata/currency-codes.txt")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		code := strings.TrimSpace(line)
		if code == "" || strings.HasPrefix(code, "#") {
			continue
		}
		for _, valid := range []bool{true, false} {
			value := CurrencyCode(code)
			if !valid {
				value = CurrencyCode("unsupported")
			}
			selector := reflect.ValueOf(Currency("currency", value)).MethodByName(code)
			if !selector.IsValid() {
				t.Fatalf("missing currency selector %s", code)
			}
			field := selector.Call(nil)[0].Interface().(Field)
			if valid {
				if err := Validate(field); err != nil {
					t.Fatalf("selector %s rejects its identifier: %v", code, err)
				}
			} else {
				failureAt(t, field, "currency")
			}
		}
	}
}

// TestCurrencyAllowlist verifies OR semantics, strict matching, and prerequisite evaluation.
func TestCurrencyAllowlist(t *testing.T) {
	for _, value := range []CurrencyCode{USD, EUR, GBP} {
		if err := Validate(Currency("currency", value).Required().USD().EUR().GBP()); err != nil {
			t.Fatal(err)
		}
	}
	for _, value := range []string{"", "usd", " USD", "US", "JPY", "unknown"} {
		failureAt(t, Currency("currency", value).USD().EUR().GBP(), "currency")
	}
	failureAt(t, Currency("currency", CurrencyCode("")).USD().Required(), "required")
	failureAt(t, Currency("currency", "").USD().RequiredIf(false), "currency")
	failureAt(t, (CurrencyField[string]{}).USD(), "currency")
	for _, field := range []Field{
		Currency("currency", USD), Currency("currency", "").Required(),
		Currency("currency", USD).USD().Using(nil), Currency("currency", USD).USD().Check(nil),
		CurrencyField[CurrencyCode]{}, (CurrencyField[CurrencyCode]{}).Message("invalid"),
	} {
		var config *ConfigurationError
		if err := Validate(field); !errors.As(err, &config) {
			t.Fatalf("invalid currency configuration accepted: %v", err)
		}
	}
	calls := 0
	var failures *Errors
	err := New().CollectAll().Validate(Currency("currency", JPY).USD().Check(func(context.Context, CurrencyCode) error { calls++; return nil }).EUR())
	if !errors.As(err, &failures) || len(failures.Issues()) != 1 || calls != 0 {
		t.Fatal(err, calls)
	}
	ctx := context.WithValue(context.Background(), struct{}{}, "request")
	field := Currency("currency", USD).USD().Using(RuleFunc[CurrencyCode](func(v CurrencyCode) error {
		if v != USD {
			t.Fatal("currency type or value changed")
		}
		return nil
	})).Check(func(received context.Context, value CurrencyCode) error {
		if received != ctx || value != USD {
			t.Fatal("context or value changed")
		}
		return nil
	})
	if err := ValidateContext(ctx, field); err != nil {
		t.Fatal(err)
	}
	op := errors.New("dependency unavailable")
	if err := Validate(Currency("currency", USD).USD().Check(func(context.Context, CurrencyCode) error { return op })); !errors.Is(err, op) {
		t.Fatal(err)
	}
}

// TestCurrencyOwnership preserves allowlist branching, custom messages, and JSON field types.
func TestCurrencyOwnership(t *testing.T) {
	base := Currency("currency", EUR).USD().Label("Billing currency").Message("Choose a supported currency")
	expanded := base.EUR()
	if err := Validate(expanded); err != nil {
		t.Fatal(err)
	}
	var failures *Errors
	if err := Validate(base); !errors.As(err, &failures) || failures.First("currency") != "Choose a supported currency" {
		t.Fatal(err)
	}
	var request struct {
		Currency CurrencyCode `json:"currency"`
	}
	if err := json.Unmarshal([]byte(`{"currency":"USD"}`), &request); err != nil || request.Currency != USD {
		t.Fatal(request, err)
	}
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := Validate(expanded.GBP()); err != nil {
				t.Error(err)
			}
			var failure *Errors
			if err := Validate(base); !errors.As(err, &failure) {
				t.Error("allowlist branch changed its base")
			}
		}()
	}
	wg.Wait()
}
