package formats

import "testing"

// TestInvalidRule verifies malformed internal identifiers fail rather than accepting input.
func TestInvalidRule(t *testing.T) {
	if err := Rule(255).Validate("anything"); err == nil {
		t.Fatal("unknown format accepted")
	}
}
