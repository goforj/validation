package is_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/goforj/validation"
	"github.com/goforj/validation/is"
)

// TestFormats verifies the catalog against positive and malformed inputs and fluent parity.
func TestFormats(t *testing.T) {
	tests := []struct {
		name string
		rule is.Rule
		good []string
		bad  []string
	}{
		{"email", is.Email, []string{"ada@example.com", "a+b@sub.example.com"}, []string{`"quoted"@example.com`, "", "Ada <ada@example.com>", " a@example.com", "a@example.com ", "a@localhost", "a@-bad.com", "a@bad-.com", "a@bad..com", "a@bad_.com", "ü@example.com", "a@例子.com", "a\n@example.com", strings.Repeat("a", 65) + "@example.com", "a@" + strings.Repeat("b", 64) + ".com", strings.Repeat("a", 255), "a@[127.0.0.1]", "bad", "a@.com"}},
		{"url", is.URL, []string{"https://example.com", "http://localhost:3000/a?q=1#result", "https://[::1]:443/", "https://127.0.0.1/", "https://example.com./"}, []string{"", "ftp://example.com", "https:///path", "https://", "https://bad host", "https://example.com:bad", "https://example.com:65536", "https://example.com:0", "https://example.com:", "https://[bad]", "https://bad_host", "https://[fe80::1%25eth0]", "https://" + strings.Repeat("a", 254), "https://-bad.com", "https://example.com/%zz"}},
		{"uuid", is.UUID, []string{"00000000-0000-0000-0000-000000000000", "ffffffff-ffff-ffff-ffff-ffffffffffff", "550E8400-E29B-41D4-A716-446655440000", "0190bde7-5ad0-7000-8000-000000000001", "12345678-1234-8123-8123-123456789abc"}, []string{"", "550e8400e29b41d4a716446655440000", "550e8400_e29b-41d4-a716-446655440000", "z50e8400-e29b-41d4-a716-446655440000", "550e8400-e29b-41d4-c716-446655440000", "550e8400-e29b-01d4-a716-446655440000", "550e8400-e29b-91d4-a716-446655440000"}},
		{"uuid_v4", is.UUIDv4, []string{"550e8400-e29b-41d4-a716-446655440000"}, []string{"", "0190bde7-5ad0-7000-8000-000000000001", "00000000-0000-0000-0000-000000000000"}},
		{"uuid_v7", is.UUIDv7, []string{"0190bde7-5ad0-7000-8000-000000000001"}, []string{"", "550e8400-e29b-41d4-a716-446655440000"}},
		{"ip", is.IP, []string{"127.0.0.1", "::1", "::ffff:192.0.2.1"}, []string{"", "localhost", "127.0.0.1:80", "192.0.2.1/24", "fe80::1%eth0"}},
		{"ipv4", is.IPv4, []string{"192.0.2.1"}, []string{"", "::1", "::ffff:192.0.2.1", "999.1.1.1"}},
		{"ipv6", is.IPv6, []string{"::1", "::ffff:192.0.2.1"}, []string{"", "192.0.2.1", "fe80::1%eth0"}},
		{"json", is.JSON, []string{`{"name":"Ada"}`, "null", "123", `"hello"`, "[]"}, []string{"", "{", "NaN", "{}{}"}},
		{"base64", is.Base64, []string{"YQ==", "YWJj"}, []string{"", "YQ", "YR==", "YQ==\n", "YQ==\r", "Y Q==", "-w=="}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for _, value := range tt.good {
				if err := tt.rule.Validate(value); err != nil {
					t.Errorf("rejected %q: %v", value, err)
				}
				if err := validation.Validate(validation.String("input", value).Is(tt.rule)); err != nil {
					t.Error(err)
				}
			}
			for _, value := range tt.bad {
				var violation *validation.Violation
				err := tt.rule.Validate(value)
				if !errors.As(err, &violation) || violation.Code() != tt.name {
					t.Errorf("%q: expected %s violation, got %v", value, tt.name, err)
					continue
				}
				var fields *validation.Errors
				if !errors.As(validation.Validate(validation.String("input", value).Is(tt.rule)), &fields) || fields.Issues()[0].Code != violation.Code() {
					t.Errorf("standalone/fluent mismatch: %q", value)
				}
			}
		})
	}
}

// TestDefinedFormatTypes ensures catalog rules do not require casting application string types.
func TestDefinedFormatTypes(t *testing.T) {
	type email string
	value := email("ada@example.com")
	if err := validation.Validate(validation.String("email", value).Is(is.Email)); err != nil {
		t.Fatal(err)
	}
	if err := is.Rule(255).Validate("anything"); err == nil {
		t.Fatal("unknown rule accepted")
	}
	if err := validation.Validate(validation.String("email", "bad").Is(is.Email).Message("Custom")); err.Error() != "Custom" {
		t.Fatal(err)
	}
}
