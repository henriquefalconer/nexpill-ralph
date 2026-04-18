// Package punycode implements RFC 3492 Punycode encoding and decoding,
// ported from punycode.js (https://github.com/mathiasbynens/punycode.js).
//
// Public surface:
//
//	Decode(input string) (string, error)      — Bootstring decoder (punycode.js:196-281)
//	Encode(input string) (string, error)      — Bootstring encoder (punycode.js:283-376)
//	ToUnicode(input string) string            — domain/email to Unicode (punycode.js:389-395)
//	ToASCII(input string) (string, error)     — domain/email to ACE (punycode.js:408-414)
//	UCS2Decode(input string) []rune           — UTF-16 decode (punycode.js:101-123)
//	UCS2Encode(codePoints []rune) string      — UTF-16 encode (punycode.js:125-133)
//
// Sentinel errors: ErrOverflow, ErrNotBasic, ErrInvalidInput.
package punycode
