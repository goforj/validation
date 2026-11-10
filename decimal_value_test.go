package validation

import (
	"math/big"
	"regexp"
	"strings"
	"testing"
)

// TestDecimalParsing checks plain-decimal syntax while preserving submitted fractional precision.
func TestDecimalParsing(t *testing.T) {
	for _, input := range []string{"0", "-0", "+0.00", "0012.3400", "-0.0001", "+123", strings.Repeat("9", 10000)} {
		if _, ok := parseDecimal(input); !ok {
			t.Errorf("rejected plain decimal: %q", input)
		}
	}
	for _, input := range []string{"", "+", "-", ".1", "1.", "1.2.3", "1e2", "1E-2", "1/2", "0x10", "NaN", "Inf", " 1", "1\n", "1,000", "1_000", "١", "１２", "\xff", "--1"} {
		if _, ok := parseDecimal(input); ok {
			t.Errorf("accepted malformed decimal: %q", input)
		}
	}
	parsed, ok := parseDecimal("-000.0000")
	if !ok || parsed.negative || parsed.integer != "" || parsed.fraction != "" || parsed.scale != 4 {
		t.Fatalf("negative zero or written scale changed: %+v", parsed)
	}
}

// TestDecimalOrdering covers magnitudes, signs, trailing zeros, and fractions beyond float precision.
func TestDecimalOrdering(t *testing.T) {
	for _, test := range []struct {
		left, right string
		want        int
	}{
		{"0", "-0.00", 0}, {"001.2300", "+1.23", 0},
		{"-1", "0", -1}, {"1", "-1", 1}, {"9", "10", -1}, {"10", "9", 1},
		{"-10", "-9", -1}, {"-9", "-10", 1}, {"10", "11", -1}, {"11", "10", 1},
		{"0.12", "0.12001", -1}, {"0.12001", "0.12", 1}, {"-0.12", "-0.12001", 1},
		{"0.09999999999999999999", "0.1", -1}, {"9007199254740993", "9007199254740992", 1},
		{"18446744073709551615.00000000000000000001", "18446744073709551615", 1},
	} {
		left, _ := parseDecimal(test.left)
		right, _ := parseDecimal(test.right)
		if got := left.compare(right); got != test.want {
			t.Errorf("compare(%s, %s) = %d, want %d", test.left, test.right, got, test.want)
		}
	}
}

// FuzzDecimalOrdering compares accepted plain decimals with an independent arbitrary-precision oracle.
func FuzzDecimalOrdering(f *testing.F) {
	plain := regexp.MustCompile(`^[+-]?[0-9]+(?:\.[0-9]+)?\z`)
	for _, pair := range [][2]string{{"0", "-0.00"}, {"001.2300", "1.23"}, {"9", "10"}, {"-0.1", "0"}, {"0.12", "0.12001"}, {"9007199254740993.01", "9007199254740993.00"}, {"1e3", "1000"}, {"", "0"}} {
		f.Add(pair[0], pair[1])
	}
	f.Fuzz(func(t *testing.T, left, right string) {
		// The independent oracle uses big integers internally; bound its test workload.
		if len(left) > 1024 || len(right) > 1024 {
			return
		}
		l, lok := parseDecimal(left)
		r, rok := parseDecimal(right)
		if lok != plain.MatchString(left) || rok != plain.MatchString(right) {
			t.Fatal("plain decimal grammar differs from reference")
		}
		if !lok || !rok {
			return
		}
		lo, lok := new(big.Rat).SetString(left)
		ro, rok := new(big.Rat).SetString(right)
		if !lok || !rok {
			t.Fatal("accepted decimal cannot be read by the reference parser")
		}
		if got, want := l.compare(r), lo.Cmp(ro); got != want {
			t.Fatalf("compare(%q, %q) = %d, want %d", left, right, got, want)
		}
	})
}

// TestDecimalComparisonResources verifies that long fractional values do not allocate magnitude-sized storage.
func TestDecimalComparisonResources(t *testing.T) {
	text := "0." + strings.Repeat("0", 1<<20) + "1"
	other, _ := parseDecimal("0.1")
	if allocations := testing.AllocsPerRun(3, func() {
		value, ok := parseDecimal(text)
		if !ok || value.compare(other) >= 0 {
			panic("incorrect exact comparison")
		}
	}); allocations != 0 {
		t.Fatalf("comparison allocated numeric storage: %v", allocations)
	}
}
