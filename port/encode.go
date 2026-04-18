package punycode

// Encode converts a Unicode string to a Punycode string.
// The "xn--" prefix is the caller's responsibility (see ToASCII).
// Mirrors punycode.js:283-376; specs/src-punycode.md:209-267.
func Encode(input string) (string, error) {
	output := make([]byte, 0, len(input)+8)

	// Convert to code points; mirrors punycode.js:294.
	codePoints := UCS2Decode(input)
	inputLength := int32(len(codePoints))

	// Initialise state. punycode.js:298-302
	n := initialN
	delta := int32(0)
	bias := initialBias

	// Copy basic code points. punycode.js:304-312
	for _, cp := range codePoints {
		if cp < 0x80 {
			output = append(output, byte(cp))
		}
	}
	basicLength := int32(len(output))
	handledCPCount := basicLength

	// Emit delimiter when there were any basic code points. punycode.js:318-320
	if basicLength > 0 {
		output = append(output, delimiter)
	}

	// Main encoding loop. punycode.js:323-374
	for handledCPCount < inputLength {

		// Find the smallest code point >= n. punycode.js:327-332
		m := maxInt
		for _, cp := range codePoints {
			if cp >= n && cp < m {
				m = cp
			}
		}

		// Overflow guard D: punycode.js:337-339
		handledCPCountPlusOne := handledCPCount + 1
		if m-n > (maxInt-delta)/handledCPCountPlusOne {
			return "", ErrOverflow
		}
		delta += (m - n) * handledCPCountPlusOne
		n = m

		for _, cp := range codePoints {
			if cp < n {
				// Overflow guard E: punycode.js:345-347
				// JS does ++delta > maxInt; we guard before increment.
				if delta >= maxInt {
					return "", ErrOverflow
				}
				delta++
			}
			if cp == n {
				// Emit variable-length integer for delta. punycode.js:350-362
				q := delta
				for k := base; ; k += base {
					t := tMin
					if k > bias+tMax {
						t = tMax
					} else if k > bias {
						t = k - bias
					}
					if q < t {
						break
					}
					qMinusT := q - t
					baseMinusT := base - t
					output = append(output, digitToBasic(t+qMinusT%baseMinusT))
					q = qMinusT / baseMinusT
				}
				// Trailing digit. punycode.js:364
				output = append(output, digitToBasic(q))

				bias = adapt(delta, handledCPCountPlusOne, handledCPCount == basicLength)
				delta = 0
				handledCPCount++
			}
		}

		delta++
		n++
	}

	return string(output), nil
}
