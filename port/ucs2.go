package punycode

// UCS2Decode converts a UTF-8 Go string into a slice of Unicode code points.
// For inputs that may contain lone UTF-16 surrogate halves, use UCS2DecodeUTF16.
// Mirrors punycode.js:101-123 for the normal (well-formed UTF-8) path.
func UCS2Decode(input string) []rune {
	output := make([]rune, 0, len(input))
	for _, r := range input {
		output = append(output, r)
	}
	return output
}

// UCS2DecodeUTF16 decodes a UTF-16 code-unit sequence into Unicode code points.
// Surrogate pairs are combined into supplementary-plane code points;
// unmatched surrogates are preserved as their raw integer values.
// Direct translation of punycode.js:101-123.
func UCS2DecodeUTF16(cu []uint16) []rune {
	output := make([]rune, 0, len(cu))
	i := 0
	for i < len(cu) {
		value := rune(cu[i])
		i++
		if value >= 0xD800 && value <= 0xDBFF && i < len(cu) {
			// High surrogate — peek at next code unit.
			extra := rune(cu[i])
			if extra&0xFC00 == 0xDC00 {
				// Valid low surrogate: combine into supplementary code point.
				output = append(output, ((value&0x3FF)<<10)+(extra&0x3FF)+0x10000)
				i++
			} else {
				// Unmatched high surrogate: emit raw; rewind so next unit is re-examined.
				output = append(output, value)
			}
		} else {
			output = append(output, value)
		}
	}
	return output
}

// UCS2Encode converts a slice of Unicode code points to a UTF-8 Go string.
// For code points that are valid Unicode scalar values this produces well-formed
// UTF-8. Use UCS2EncodeUTF16 when lone-surrogate round-trip fidelity is required.
// Mirrors punycode.js:125-133.
func UCS2Encode(codePoints []rune) string {
	return string(codePoints)
}

// UCS2EncodeUTF16 converts code points to a UTF-16 code-unit sequence.
// Values > 0xFFFF become high/low surrogate pairs; values in 0xD800..0xDFFF
// pass through unchanged as single uint16 values (lone-surrogate preservation).
func UCS2EncodeUTF16(codePoints []rune) []uint16 {
	cu := make([]uint16, 0, len(codePoints))
	for _, cp := range codePoints {
		if cp > 0xFFFF {
			cp -= 0x10000
			cu = append(cu, uint16(0xD800+(cp>>10)))
			cu = append(cu, uint16(0xDC00+(cp&0x3FF)))
		} else {
			cu = append(cu, uint16(cp))
		}
	}
	return cu
}
