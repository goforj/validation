package validation

import "strings"

// decimalValue separates significant digits from written scale for exact comparisons.
type decimalValue struct {
	negative bool
	integer  string
	fraction string
	scale    int
}

// parseDecimal retains written scale while extracting significant digits from plain ASCII decimals.
func parseDecimal(value string) (decimalValue, bool) {
	var parsed decimalValue
	if value == "" {
		return parsed, false
	}
	if value[0] == '-' || value[0] == '+' {
		parsed.negative = value[0] == '-'
		value = value[1:]
	}
	integer, fraction, point := strings.Cut(value, ".")
	if integer == "" || point && fraction == "" {
		return decimalValue{}, false
	}
	for _, digits := range []string{integer, fraction} {
		for i := 0; i < len(digits); i++ {
			if digits[i] < '0' || digits[i] > '9' {
				return decimalValue{}, false
			}
		}
	}
	parsed.integer = strings.TrimLeft(integer, "0")
	parsed.fraction = strings.TrimRight(fraction, "0")
	parsed.scale = len(fraction)
	if parsed.integer == "" && parsed.fraction == "" {
		parsed.negative = false
	}
	return parsed, true
}

// compare orders decimal values without floating-point rounding or magnitude-sized allocations.
// Trimming trailing fractional zeros makes an equal prefix order numerically (0.12 < 0.1201).
func (value decimalValue) compare(other decimalValue) int {
	if value.negative != other.negative {
		if value.negative {
			return -1
		}
		return 1
	}
	order := 0
	if len(value.integer) < len(other.integer) {
		order = -1
	} else if len(value.integer) > len(other.integer) {
		order = 1
	} else {
		order = strings.Compare(value.integer, other.integer)
		if order == 0 {
			order = strings.Compare(value.fraction, other.fraction)
		}
	}
	if value.negative {
		return -order
	}
	return order
}
