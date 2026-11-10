// Package is supplies immutable, reusable string-format validation rules.
// Rules work independently through Validate and inside validation.String.Is.
// Every rule rejects empty input. Checks are local and never access the network.
package is

import "github.com/goforj/validation/internal/formats"

// Rule is an immutable format identifier implementing validation.Rule[string].
// Use the named constants; other numeric identifiers are invalid configuration.
type Rule = formats.Rule

const (
	// Email requires a bare ASCII mailbox, a dotted DNS domain, and no whitespace padding.
	Email Rule = formats.Email
	// URL requires an absolute HTTP or HTTPS URL with an ASCII DNS host or IP literal.
	URL Rule = formats.URL
	// UUID accepts canonical UUID versions 1 through 8, plus nil and max UUIDs.
	UUID Rule = formats.UUID
	// UUIDv4 requires a canonical version 4 UUID with the RFC variant.
	UUIDv4 Rule = formats.UUIDv4
	// UUIDv7 requires a canonical version 7 UUID with the RFC variant.
	UUIDv7 Rule = formats.UUIDv7
	// IP accepts unzoned IP literals without ports or CIDR prefixes.
	IP Rule = formats.IP
	// IPv4 accepts IPv4 literals without ports or CIDR prefixes.
	IPv4 Rule = formats.IPv4
	// IPv6 accepts unzoned IPv6 literals, including IPv4-mapped IPv6 addresses.
	IPv6 Rule = formats.IPv6
	// JSON accepts any valid JSON document, including null and scalar values.
	JSON Rule = formats.JSON
	// Base64 accepts nonempty strict standard padded Base64 without whitespace.
	Base64 Rule = formats.Base64
)
