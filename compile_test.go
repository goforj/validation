package validation

import (
	"os/exec"
	"strings"
	"testing"
)

// TestInvalidAPIsDoNotCompile protects typed rule families and exact bound types.
func TestInvalidAPIsDoNotCompile(t *testing.T) {
	command := exec.Command("go", "test", "./testdata/compilefail")
	output, err := command.CombinedOutput()
	if err == nil {
		t.Fatal("invalid API usage compiled")
	}
	for _, expected := range []string{"has no field or method Email", "1.5", "has no field or method Required", "Rule[string]"} {
		if !strings.Contains(string(output), expected) {
			t.Fatalf("missing %q in compiler diagnostics:\n%s", expected, output)
		}
	}
}
