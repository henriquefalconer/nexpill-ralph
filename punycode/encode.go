package punycode

import "strings"

// Encode converts a Unicode string to its Punycode representation.
// The xn-- ACE prefix is not prepended; that is ToASCII's responsibility.
// Mirrors punycode.js:290-376 (RFC 3492 §6.3).
func Encode(input string) (string, error) {
	codePoints := UCS2Decode(input)
	inputLength := len(codePoints)

	var sb strings.Builder

	// Step 1: copy basic code points (< U+0080) to output.
	for _, cp := range codePoints {
		if cp < 0x80 {
			sb.WriteRune(cp)
		}
	}

	basicLength := sb.Len() // byte count == rune count for ASCII
	handledCPCount := basicLength

	// Step 2: emit the Punycode delimiter only when basic section is non-empty.
	if basicLength > 0 {
		sb.WriteRune(delimiter)
	}

	// Step 3: main encoding loop — process non-basic code points.
	n := initialN
	delta := 0
	bias := initialBias

	for handledCPCount < inputLength {
		// 3a. Find the smallest code point m >= n not yet handled.
		m := maxInt
		for _, cp := range codePoints {
			v := int(cp)
			if v >= n && v < m {
				m = v
			}
		}

		// 3b. Advance delta by (m-n)*(handledCPCount+1), guarding overflow.
		handledCPCountPlusOne := handledCPCount + 1
		if m-n > (maxInt-delta)/handledCPCountPlusOne {
			return "", ErrOverflow
		}
		delta += (m - n) * handledCPCountPlusOne
		n = m

		// 3c. Inner pass: increment for code points < n; emit VLQ for == n.
		for _, cp := range codePoints {
			v := int(cp)
			if v < n {
				delta++
				if delta > maxInt {
					return "", ErrOverflow
				}
			}
			if v == n {
				// Encode delta as a generalized variable-length integer.
				q := delta
				for k := base; ; k += base {
					t := tMin
					if k > bias {
						if k >= bias+tMax {
							t = tMax
						} else {
							t = k - bias
						}
					}
					if q < t {
						break
					}
					qMinusT := q - t
					baseMinusT := base - t
					sb.WriteRune(digitToBasic(t+qMinusT%baseMinusT, false))
					q = qMinusT / baseMinusT
				}
				sb.WriteRune(digitToBasic(q, false))
				bias = adapt(delta, handledCPCountPlusOne, handledCPCount == basicLength)
				delta = 0
				handledCPCount++
			}
		}

		// 3d. Advance past this code point.
		delta++
		n++
	}

	return sb.String(), nil
}
