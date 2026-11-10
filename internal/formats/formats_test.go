package formats

import (
	"encoding/base64"
	"strings"
	"testing"
)

// TestInvalidRule verifies malformed internal identifiers fail rather than accepting input.
func TestInvalidRule(t *testing.T) {
	if err := Rule(255).Validate("anything"); err == nil {
		t.Fatal("unknown format accepted")
	}
}

// TestBase64Blocks verifies strict padding and trailing bits across scratch-buffer boundaries.
func TestBase64Blocks(t *testing.T) {
	for _, size := range []int{1, 2, 3, 767, 768, 769, 1536, 4096} {
		value := base64.StdEncoding.EncodeToString(make([]byte, size))
		if !validBase64(value) {
			t.Fatalf("rejected %d decoded bytes", size)
		}
	}
	for _, value := range []string{
		strings.Repeat("YWJj", 255) + "YQ==" + "YWJj",
		strings.Repeat("YWJj", 256) + "YR==",
		"!" + strings.Repeat("YWJj", 256) + "abc",
		strings.Repeat("YWJj", 256) + "YWJ!",
	} {
		if validBase64(value) {
			t.Fatal("accepted malformed Base64 across a block boundary")
		}
	}
	value := strings.Repeat("YWJj", 1<<20)
	if allocations := testing.AllocsPerRun(3, func() {
		if !validBase64(value) {
			panic("valid payload rejected")
		}
	}); allocations != 0 {
		t.Fatalf("validation allocated decoded storage: %v allocations", allocations)
	}
}

// FuzzBase64 compares bounded validation against the standard decoder's complete-input result.
func FuzzBase64(f *testing.F) {
	for _, value := range []string{"", "YQ==", "YR==", "YWJj", "YQ==\n", strings.Repeat("YWJj", 255) + "YQ==YWJj", strings.Repeat("YWJj", 256) + "YQ=="} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		_, err := base64.StdEncoding.Strict().DecodeString(value)
		want := value != "" && err == nil && !strings.ContainsAny(value, "\r\n")
		if got := validBase64(value); got != want {
			t.Fatalf("validity differs: got %v, want %v", got, want)
		}
	})
}

// FuzzFormats exercises every format against arbitrary bytes, including malformed UTF-8.
func FuzzFormats(f *testing.F) {
	for _, value := range []string{"", "ada@example.com", "https://[::1]:443/?q=%20", "https://::1:443", "https://example.com/?q=%zz", "0190bde7-5ad0-7000-8000-000000000001", "{\"ok\":true}", "YQ==", "\xff\x00%"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		for rule := Email; rule <= UTF8; rule++ {
			if err := rule.Validate(value); err != nil && err.Error() == "" {
				t.Fatal("format returned an empty failure")
			}
		}
	})
}

// FuzzBase64URL compares chunked validation with the complete standard-library decoder.
func FuzzBase64URL(f *testing.F) {
	for _, value := range []string{"", "YQ", "YQ==", "YR", "__8", "__8=", "__9=", strings.Repeat("YWJj", 257) + "YQ"} {
		f.Add(value)
	}
	f.Fuzz(func(t *testing.T, value string) {
		decoder := base64.RawURLEncoding.Strict()
		if strings.ContainsRune(value, '=') {
			decoder = base64.URLEncoding.Strict()
		}
		_, err := decoder.DecodeString(value)
		want := value != "" && err == nil && !strings.ContainsAny(value, "\r\n")
		if got := validBase64URL(value); got != want {
			t.Fatalf("got %v, want %v", got, want)
		}
	})
}
