package validation

import "testing"

// TestAdditionalFormats checks character, identifier, network, and encoding boundaries.
func TestAdditionalFormats(t *testing.T) {
	for _, tc := range []struct {
		name      string
		build     func(string) Field
		good, bad []string
	}{
		{"ulid", func(v string) Field { return String("v", v).IsULID() }, []string{"01ARZ3NDEKTSV4RRFFQ69G5FAV", "7zzzzzzzzzzzzzzzzzzzzzzzzz", "00000000000000000000000000"}, []string{"", "8ZZZZZZZZZZZZZZZZZZZZZZZZZ", "01ARZ3NDEKTSV4RRFFQ69G5FAI", "/1ARZ3NDEKTSV4RRFFQ69G5FAV"}},
		{"hostname", func(v string) Field { return String("v", v).IsHostname() }, []string{"localhost", "example.com.", "xn--bcher-kva.example"}, []string{"", ".", "bad..host", "-host", "é.example", "host:80"}},
		{"cidr", func(v string) Field { return String("v", v).IsCIDR() }, []string{"192.0.2.1/24", "2001:db8::/32"}, []string{"", "192.0.2.1/33", "fe80::1%eth0/64", "::1"}},
		{"mac", func(v string) Field { return String("v", v).IsMAC() }, []string{"00:11:22:33:44:55", "00-11-22-33-44-55-66-77", "0011.2233.4455"}, []string{"", "00:11:22", "00:11:22:33:44:55:66:77:88:99:aa:bb:cc:dd:ee:ff:00:11:22:33"}},
		{"e164", func(v string) Field { return String("v", v).IsE164() }, []string{"+12", "+14155552671", "+123456789012345"}, []string{"", "+1", "+012", "123", "+1 23", "+1234567890123456", "+1é"}},
		{"base64_url", func(v string) Field { return String("v", v).IsBase64URL() }, []string{"__8", "__8=", "YQ", "YQ=="}, []string{"", "Y", "YR", "__9=", "//8=", "YQ\n", "YQ==YQ=="}},
		{"alpha", func(v string) Field { return String("v", v).IsAlpha() }, []string{"Abc"}, []string{"", "abc1", "é"}},
		{"alphanumeric", func(v string) Field { return String("v", v).IsAlphanumeric() }, []string{"Abc123", "123"}, []string{"", "a-b", "é"}},
		{"digits", func(v string) Field { return String("v", v).IsDigits() }, []string{"0123"}, []string{"", "1.2", "١", "a"}},
		{"ascii", func(v string) Field { return String("v", v).IsASCII() }, []string{"abc123\x00\n"}, []string{"", "é", "\xff"}},
		{"lowercase", func(v string) Field { return String("v", v).IsLowercase() }, []string{"école 123", "hello 世界"}, []string{"", "123世界", "ABC", "ǅ", "\xff"}},
		{"uppercase", func(v string) Field { return String("v", v).IsUppercase() }, []string{"ÉCOLE 123", "HELLO 世界"}, []string{"", "123世界", "abc", "ǅ", "\xff"}},
		{"utf8", func(v string) Field { return String("v", v).IsUTF8() }, []string{"", "世界", "�"}, []string{"\xff", "\xc0\x80"}},
	} {
		for _, v := range tc.good {
			if err := Validate(tc.build(v)); err != nil {
				t.Errorf("%s rejected %q: %v", tc.name, v, err)
			}
		}
		for _, v := range tc.bad {
			failureAt(t, tc.build(v), tc.name)
		}
	}
}
