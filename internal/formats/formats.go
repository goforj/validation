// Package formats owns the checks behind fluent string format methods.
package formats

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net"
	"net/mail"
	"net/netip"
	"net/url"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/goforj/validation/internal/violation"
)

// Rule identifies an immutable string-format rule.
type Rule uint8

const (
	// Email checks a bare ASCII mailbox without DNS access.
	Email Rule = iota + 1
	// URL checks an absolute HTTP or HTTPS URL.
	URL
	// UUID checks a canonical UUID, including nil and max UUIDs.
	UUID
	// UUIDv4 checks a canonical version 4 UUID.
	UUIDv4
	// UUIDv7 checks a canonical version 7 UUID.
	UUIDv7
	// IP checks an IP address literal without a zone or prefix.
	IP
	// IPv4 checks an IPv4 address literal.
	IPv4
	// IPv6 checks an IPv6 address literal, including mapped IPv4 addresses.
	IPv6
	// JSON checks any nonempty valid JSON document, including scalar values.
	JSON
	// Base64 checks strict standard padded Base64 without whitespace.
	Base64
	// ULID checks a 128-bit Crockford Base32 identifier.
	ULID
	// Hostname checks ASCII DNS labels without resolution.
	Hostname
	// CIDR checks an IP prefix without requiring masked host bits.
	CIDR
	// MAC checks a 48-bit or 64-bit link-layer address.
	MAC
	// E164 checks international phone-number syntax without dialing metadata.
	E164
	// Base64URL checks padded or raw URL-safe Base64.
	Base64URL
	// Alpha checks nonempty ASCII letters.
	Alpha
	// Alphanumeric checks nonempty ASCII letters and digits.
	Alphanumeric
	// Digits checks nonempty ASCII decimal digits.
	Digits
	// ASCII checks nonempty ASCII text, including control bytes.
	ASCII
	// Lowercase checks valid UTF-8 with lowercase cased letters.
	Lowercase
	// Uppercase checks valid UTF-8 with uppercase cased letters.
	Uppercase
	// UTF8 checks valid UTF-8, including empty text.
	UTF8
)

// Validate returns a coded failure without retaining the submitted string.
func (r Rule) Validate(value string) error {
	var valid bool
	var code, message string
	switch r {
	case Email:
		valid, code, message = email(value), "email", "must be a valid email address"
	case URL:
		valid, code, message = webURL(value), "url", "must be an absolute HTTP or HTTPS URL"
	case UUID, UUIDv4, UUIDv7:
		valid, code, message = uuid(value, r), "uuid", "must be a valid UUID"
		if r == UUIDv4 {
			code, message = "uuid_v4", "must be a valid UUID v4"
		}
		if r == UUIDv7 {
			code, message = "uuid_v7", "must be a valid UUID v7"
		}
	case IP, IPv4, IPv6:
		address, err := netip.ParseAddr(value)
		valid = err == nil && address.Zone() == ""
		code, message = "ip", "must be a valid IP address"
		if r == IPv4 {
			valid = valid && address.Is4()
			code, message = "ipv4", "must be a valid IPv4 address"
		}
		if r == IPv6 {
			valid = valid && address.Is6()
			code, message = "ipv6", "must be a valid IPv6 address"
		}
	case JSON:
		valid, code, message = json.Valid([]byte(value)), "json", "must be valid JSON"
	case Base64:
		valid, code, message = validBase64(value), "base64", "must be standard padded Base64"
	case ULID:
		valid, code, message = ulid(value), "ulid", "must be a valid ULID"
	case Hostname:
		valid, code, message = hostname(strings.TrimSuffix(value, ".")), "hostname", "must be a valid ASCII hostname"
	case CIDR:
		_, err := netip.ParsePrefix(value)
		valid, code, message = err == nil, "cidr", "must be a valid IP prefix"
	case MAC:
		address, err := net.ParseMAC(value)
		valid, code, message = err == nil && (len(address) == 6 || len(address) == 8), "mac", "must be a valid MAC address"
	case E164:
		valid, code, message = phone(value), "e164", "must be an international phone number"
	case Base64URL:
		valid, code, message = validBase64URL(value), "base64_url", "must be URL-safe Base64"
	case Alpha, Alphanumeric, Digits, ASCII:
		valid = asciiText(value, r)
		switch r {
		case Alpha:
			code, message = "alpha", "must contain only ASCII letters"
		case Alphanumeric:
			code, message = "alphanumeric", "must contain only ASCII letters and digits"
		case Digits:
			code, message = "digits", "must contain only ASCII digits"
		case ASCII:
			code, message = "ascii", "must contain only ASCII characters"
		}
	case Lowercase, Uppercase:
		valid = letterCase(value, r)
		code, message = "lowercase", "must use lowercase letters"
		if r == Uppercase {
			code, message = "uppercase", "must use uppercase letters"
		}
	case UTF8:
		valid, code, message = utf8.ValidString(value), "utf8", "must be valid UTF-8"
	default:
		return fmt.Errorf("validation: unknown format rule %d", r)
	}
	if valid {
		return nil
	}
	return violation.Builtin(code, message, nil)
}

// validBase64 bounds scratch memory while retaining the standard decoder's strict padding checks.
func validBase64(value string) bool {
	if value == "" || len(value)%4 != 0 || strings.ContainsAny(value, "\r\n") {
		return false
	}
	return decodeBase64(value, base64.StdEncoding.Strict())
}

// decodeBase64 bounds scratch storage for both standard and URL-safe alphabets.
func decodeBase64(value string, decoder *base64.Encoding) bool {
	const block = 1024
	var scratch [block / 4 * 3]byte
	for len(value) > block {
		chunk := value[:block]
		// Padding is legal only in the final quantum of the complete input.
		if strings.ContainsRune(chunk, '=') {
			return false
		}
		if _, err := decoder.Decode(scratch[:], []byte(chunk)); err != nil {
			return false
		}
		value = value[block:]
	}
	_, err := decoder.Decode(scratch[:], []byte(value))
	return err == nil
}

// email enforces the application's ASCII mailbox policy in addition to mail syntax.
func email(value string) bool {
	if len(value) > 254 || strings.TrimSpace(value) != value {
		return false
	}
	for _, r := range value {
		if r > unicode.MaxASCII || unicode.IsControl(r) {
			return false
		}
	}
	address, err := mail.ParseAddress(value)
	if err != nil || address.Name != "" || address.Address != value {
		return false
	}
	at := strings.LastIndexByte(value, '@')
	if at < 1 || at > 64 {
		return false
	}
	domain := value[at+1:]
	return strings.Contains(domain, ".") && hostname(domain)
}

// hostname rejects malformed ASCII DNS labels without making network requests.
func hostname(value string) bool {
	if value == "" || len(value) > 253 {
		return false
	}
	for _, label := range strings.Split(value, ".") {
		if len(label) == 0 || len(label) > 63 || label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for _, r := range label {
			if !(r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-') {
				return false
			}
		}
	}
	return true
}

// webURL enforces explicit web schemes, valid hosts, ports, and whitespace-free spelling.
func webURL(value string) bool {
	if strings.ContainsFunc(value, unicode.IsSpace) || !validEscapes(value) {
		return false
	}
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return false
	}
	host := u.Hostname()
	address, ipError := netip.ParseAddr(host)
	if ipError != nil {
		if strings.Contains(u.Host, "[") || !hostname(strings.TrimSuffix(host, ".")) {
			return false
		}
	} else if address.Zone() != "" || address.Is6() != strings.HasPrefix(u.Host, "[") {
		return false
	}
	if strings.HasSuffix(u.Host, ":") {
		return false
	}
	if port := u.Port(); port != "" {
		n, err := strconv.ParseUint(port, 10, 16)
		if err != nil || n == 0 {
			return false
		}
	}
	return true
}

// validEscapes checks query escapes that net/url deliberately leaves uninterpreted.
func validEscapes(value string) bool {
	for i := 0; i < len(value); i++ {
		if value[i] != '%' {
			continue
		}
		if i+2 >= len(value) || !hexDigit(value[i+1]) || !hexDigit(value[i+2]) {
			return false
		}
		i += 2
	}
	return true
}

// hexDigit recognizes URI escape digits without allocating a decoded component.
func hexDigit(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F'
}

// uuid checks canonical spelling and the RFC variant without allocating decoded bytes.
func uuid(value string, rule Rule) bool {
	if len(value) != 36 {
		return false
	}
	value = strings.ToLower(value)
	for i, r := range value {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if r != '-' {
				return false
			}
		} else if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f') {
			return false
		}
	}
	if rule == UUID && (value == "00000000-0000-0000-0000-000000000000" || value == "ffffffff-ffff-ffff-ffff-ffffffffffff") {
		return true
	}
	if value[19] < '8' || value[19] > 'b' {
		return false
	}
	switch rule {
	case UUIDv4:
		return value[14] == '4'
	case UUIDv7:
		return value[14] == '7'
	default:
		return value[14] >= '1' && value[14] <= '8'
	}
}

// validBase64URL accepts either complete padded or canonical unpadded URL-safe input.
func validBase64URL(value string) bool {
	if value == "" || strings.ContainsAny(value, "\r\n") {
		return false
	}
	decoder := base64.RawURLEncoding.Strict()
	if strings.ContainsRune(value, '=') {
		decoder = base64.URLEncoding.Strict()
	}
	return decodeBase64(value, decoder)
}

// ulid enforces the specification's alphabet and 128-bit overflow limit.
func ulid(value string) bool {
	if len(value) != 26 || value[0] < '0' || value[0] > '7' {
		return false
	}
	for _, r := range value {
		if !strings.ContainsRune("0123456789ABCDEFGHJKMNPQRSTVWXYZabcdefghjkmnpqrstvwxyz", r) {
			return false
		}
	}
	return true
}

// phone checks E.164 spelling without claiming assignment or reachability.
func phone(value string) bool {
	if len(value) < 3 || len(value) > 16 || value[0] != '+' || value[1] < '1' || value[1] > '9' {
		return false
	}
	for _, r := range value[2:] {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// asciiText keeps character-class policies explicit and independent of Unicode locale.
func asciiText(value string, rule Rule) bool {
	if value == "" {
		return false
	}
	for _, r := range value {
		letter := r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
		digit := r >= '0' && r <= '9'
		if rule == Alpha && !letter || rule == Alphanumeric && !letter && !digit || rule == Digits && !digit || r > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// letterCase requires a cased letter and rejects malformed UTF-8 and opposite case.
func letterCase(value string, rule Rule) bool {
	if !utf8.ValidString(value) {
		return false
	}
	found := false
	for _, r := range value {
		lower, upper := unicode.IsLower(r), unicode.IsUpper(r)
		if unicode.IsTitle(r) || rule == Lowercase && upper || rule == Uppercase && lower {
			return false
		}
		found = found || lower || upper
	}
	return found
}
