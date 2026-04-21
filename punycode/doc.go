// Package punycode implements the Punycode encoding algorithm (RFC 3492) and
// Internationalizing Domain Names in Applications (IDNA, RFC 5891).
//
// It is a faithful Go port of mathiasbynens/punycode.js version 2.3.1
// (https://github.com/mathiasbynens/punycode.js). The public API mirrors the
// JS library with idiomatic Go adjustments:
//
//   - The JS ucs2.decode / ucs2.encode nested namespace is flattened to
//     [UCS2Decode] / [UCS2DecodeUnits] and [UCS2Encode].
//   - All fallible functions return (string, error) rather than throwing.
//   - Three sentinel error values ([ErrOverflow], [ErrNotBasic],
//     [ErrInvalidInput]) correspond to the JS errors object (punycode.js:22-25).
//
// # Public surface
//
//	Version       — library version string ("2.3.1")
//	Encode        — Unicode string → Punycode label (no "xn--" prefix)
//	Decode        — Punycode label → Unicode string
//	ToASCII       — Unicode domain/email → ACE form (adds "xn--" per label)
//	ToUnicode     — ACE domain/email → Unicode form
//	UCS2Decode    — UTF-8 string → []rune (well-formed UTF-8)
//	UCS2DecodeUnits — []uint16 UTF-16 units → []rune (surrogate-aware)
//	UCS2Encode    — []rune → UTF-8 string (WTF-8 for surrogate code points)
//
// # Quick start
//
//	// Encode a Unicode label (without the "xn--" ACE prefix):
//	enc, err := punycode.Encode("München")
//
//	// Convert a full Unicode domain to its ASCII-compatible encoding:
//	ascii, err := punycode.ToASCII("www.münchen.de")
//
//	// Decode back to Unicode:
//	uni, err := punycode.ToUnicode("www.xn--mnchen-3ya.de")
//
// # References
//
//   - RFC 3492 — Punycode: A Bootstring encoding of Unicode for IDNA
//     (https://www.rfc-editor.org/rfc/rfc3492)
//   - RFC 5891 — Internationalized Domain Names in Applications (IDNA): Protocol
//     (https://www.rfc-editor.org/rfc/rfc5891)
//   - Upstream JS library: https://github.com/mathiasbynens/punycode.js
package punycode
