package validation

import "github.com/goforj/validation/internal/formats"

// IsULID requires a case-insensitive 128-bit ULID.
func (b StringField[T]) IsULID() StringField[T] { return b.format(formats.ULID) }

// IsHostname requires ASCII DNS labels, optionally ending in a root dot.
func (b StringField[T]) IsHostname() StringField[T] { return b.format(formats.Hostname) }

// IsCIDR requires an IP prefix without normalizing host bits.
func (b StringField[T]) IsCIDR() StringField[T] { return b.format(formats.CIDR) }

// IsMAC requires a 48-bit or 64-bit link-layer address.
func (b StringField[T]) IsMAC() StringField[T] { return b.format(formats.MAC) }

// IsE164 requires international phone-number syntax, without assignment checks.
func (b StringField[T]) IsE164() StringField[T] { return b.format(formats.E164) }

// IsBase64URL requires strict padded or raw URL-safe Base64.
func (b StringField[T]) IsBase64URL() StringField[T] { return b.format(formats.Base64URL) }

// IsAlpha requires nonempty ASCII letters.
func (b StringField[T]) IsAlpha() StringField[T] { return b.format(formats.Alpha) }

// IsAlphanumeric requires nonempty ASCII letters and digits.
func (b StringField[T]) IsAlphanumeric() StringField[T] { return b.format(formats.Alphanumeric) }

// IsDigits requires nonempty ASCII decimal digits.
func (b StringField[T]) IsDigits() StringField[T] { return b.format(formats.Digits) }

// IsASCII requires nonempty ASCII text, including control bytes.
func (b StringField[T]) IsASCII() StringField[T] { return b.format(formats.ASCII) }

// IsLowercase requires valid UTF-8 with at least one cased letter and no uppercase or titlecase letters.
func (b StringField[T]) IsLowercase() StringField[T] { return b.format(formats.Lowercase) }

// IsUppercase requires valid UTF-8 with at least one cased letter and no lowercase or titlecase letters.
func (b StringField[T]) IsUppercase() StringField[T] { return b.format(formats.Uppercase) }

// IsUTF8 requires valid UTF-8, including empty text.
func (b StringField[T]) IsUTF8() StringField[T] { return b.format(formats.UTF8) }
