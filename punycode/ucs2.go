package punycode

import "strings"

// UCS2Decode converts a Go UTF-8 string to a slice of Unicode code points.
// For well-formed UTF-8 input this is equivalent to []rune(s).
// Use UCS2DecodeUnits for surrogate-aware decoding of raw UTF-16 code units.
func UCS2Decode(s string) []rune {
	return []rune(s)
}

// UCS2DecodeUnits converts a UTF-16 code unit sequence to Unicode code points.
// Surrogate pairs are combined; unmatched surrogates are emitted as-is.
// Mirrors the JS ucs2.decode algorithm (punycode.js:101-123).
func UCS2DecodeUnits(units []uint16) []rune {
	out := make([]rune, 0, len(units))
	for i := 0; i < len(units); {
		high := rune(units[i])
		i++
		if high >= 0xD800 && high <= 0xDBFF && i < len(units) {
			low := rune(units[i])
			if low >= 0xDC00 && low <= 0xDFFF {
				high = ((high & 0x3FF) << 10) + (low & 0x3FF) + 0x10000
				i++
			}
			// lone high surrogate: emit as-is, continue from next unit
		}
		out = append(out, high)
	}
	return out
}

// UCS2Encode converts a slice of Unicode code points to a UTF-8 string.
// Surrogate code points (U+D800–U+DFFF) are encoded as WTF-8 (3-byte sequences)
// so that round-trips through UCS2DecodeUnits are lossless.
// Mirrors punycode.js:133 (String.fromCodePoint).
func UCS2Encode(codePoints []rune) string {
	var b strings.Builder
	b.Grow(len(codePoints))
	for _, cp := range codePoints {
		if cp >= 0xD800 && cp <= 0xDFFF {
			// WTF-8: encode surrogate as 3-byte sequence
			b.WriteByte(byte(0xE0 | (cp >> 12)))
			b.WriteByte(byte(0x80 | ((cp >> 6) & 0x3F)))
			b.WriteByte(byte(0x80 | (cp & 0x3F)))
		} else {
			b.WriteRune(cp)
		}
	}
	return b.String()
}
