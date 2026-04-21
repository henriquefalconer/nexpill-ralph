package punycode

import "strings"

// Decode converts a Punycode-encoded string to Unicode.
// Mirrors punycode.js:196-281 (RFC 3492 §6.2).
func Decode(input string) (string, error) {
	output := make([]rune, 0, len(input))
	inputLength := len(input)
	i := 0
	n := initialN
	bias := initialBias

	// Locate the last delimiter to separate the basic prefix.
	basic := strings.LastIndex(input, string(delimiter))
	if basic < 0 {
		basic = 0
	}

	// Copy basic code points from the prefix verbatim.
	for j := 0; j < basic; j++ {
		if input[j] >= 0x80 {
			return "", ErrNotBasic
		}
		output = append(output, rune(input[j]))
	}

	// Start past the delimiter if there were any basic code points.
	startIndex := 0
	if basic > 0 {
		startIndex = basic + 1
	}

	for index := startIndex; index < inputLength; {
		oldi := i

		// Decode one variable-length integer (delta).
		for w, k := 1, base; ; k += base {
			if index >= inputLength {
				return "", ErrInvalidInput
			}
			digit := basicToDigit(rune(input[index]))
			index++
			if digit >= base {
				return "", ErrInvalidInput
			}
			if digit > (maxInt-i)/w {
				return "", ErrOverflow
			}
			i += digit * w

			t := tMin
			if k > bias {
				if k >= bias+tMax {
					t = tMax
				} else {
					t = k - bias
				}
			}
			if digit < t {
				break
			}

			baseMinusT := base - t
			if w > maxInt/baseMinusT {
				return "", ErrOverflow
			}
			w *= baseMinusT
		}

		out := len(output) + 1
		bias = adapt(i-oldi, out, oldi == 0)

		if i/out > maxInt-n {
			return "", ErrOverflow
		}
		n += i / out
		i %= out

		// Insert rune(n) at position i (mirrors output.splice(i, 0, n)).
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = rune(n)
		i++
	}

	return string(output), nil
}
