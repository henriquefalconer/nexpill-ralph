package punycode

import "strings"

// Decode converts a Punycode string to a Unicode string.
// The "xn--" prefix is the caller's responsibility (see ToUnicode).
// Mirrors punycode.js:196-281; specs/src-punycode.md:152-207.
func Decode(input string) (string, error) {
	output := make([]rune, 0, len(input))
	i := int32(0)
	n := initialN
	bias := initialBias

	// Find the last delimiter to split basic from non-basic code points.
	// punycode.js:208-219
	basic := strings.LastIndexByte(input, '-')
	if basic < 0 {
		basic = 0
	}

	for j := 0; j < basic; j++ {
		if input[j] >= 0x80 {
			return "", ErrNotBasic
		}
		output = append(output, rune(input[j]))
	}

	// Main decode loop. punycode.js:224-278
	index := 0
	if basic > 0 {
		index = basic + 1
	}
	inputLength := len(input)

	for index < inputLength {
		oldi := i

		// Decode a generalised variable-length integer. punycode.js:231-261
		w := int32(1)
		for k := base; ; k += base {
			if index >= inputLength {
				return "", ErrInvalidInput
			}
			digit := basicToDigit(int32(input[index]))
			index++

			if digit >= base {
				return "", ErrInvalidInput
			}
			// Overflow guard A: punycode.js:243-245
			if digit > (maxInt-i)/w {
				return "", ErrOverflow
			}
			i += digit * w

			t := tMin
			if k > bias+tMax {
				t = tMax
			} else if k > bias {
				t = k - bias
			}

			if digit < t {
				break
			}

			baseMinusT := base - t
			// Overflow guard B: punycode.js:255-257
			if w > maxInt/baseMinusT {
				return "", ErrOverflow
			}
			w *= baseMinusT
		}

		out := int32(len(output)) + 1
		bias = adapt(i-oldi, out, oldi == 0)

		// Overflow guard C: punycode.js:268-270
		if i/out > maxInt-n {
			return "", ErrOverflow
		}
		n += i / out
		i %= out

		// Insert n at position i. punycode.js:278
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = rune(n)
		i++
	}

	return UCS2Encode(output), nil
}
