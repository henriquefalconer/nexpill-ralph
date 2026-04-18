package punycode

import (
	"errors"
	"strings"
)

const (
	maxInt        int32 = 2147483647
	base          int32 = 36
	tMin          int32 = 1
	tMax          int32 = 26
	skew          int32 = 38
	damp          int32 = 700
	initialBias   int32 = 72
	initialN      int32 = 128
	delim               = '-'
	baseMinusTMin int32 = base - tMin
)

var (
	ErrOverflow     = errors.New("Overflow: input needs wider integers to process")
	ErrNotBasic     = errors.New("Illegal input >= 0x80 (not a basic code point)")
	ErrInvalidInput = errors.New("Invalid input")
)

// UCS2Decode converts a UTF-16 code unit sequence to Unicode code points.
// Surrogate pairs are combined; lone surrogates are preserved as-is.
func UCS2Decode(input []uint16) []int32 {
	out := make([]int32, 0, len(input))
	for i := 0; i < len(input); {
		v := input[i]
		i++
		if v >= 0xD800 && v <= 0xDBFF && i < len(input) {
			extra := input[i]
			if extra >= 0xDC00 && extra <= 0xDFFF {
				i++
				cp := int32(((uint32(v)&0x3FF)<<10)|(uint32(extra)&0x3FF)) + 0x10000
				out = append(out, cp)
				continue
			}
		}
		out = append(out, int32(v))
	}
	return out
}

// UCS2Encode converts Unicode code points to a UTF-16 code unit sequence.
// Code points > U+FFFF become surrogate pairs; lone surrogates pass through.
func UCS2Encode(codePoints []int32) []uint16 {
	out := make([]uint16, 0, len(codePoints))
	for _, cp := range codePoints {
		if cp > 0xFFFF {
			cp -= 0x10000
			out = append(out, uint16(0xD800+(cp>>10)), uint16(0xDC00+(cp&0x3FF)))
		} else {
			out = append(out, uint16(cp))
		}
	}
	return out
}

func basicToDigit(cp int32) int32 {
	if cp >= 0x30 && cp < 0x3A {
		return 26 + (cp - 0x30)
	}
	if cp >= 0x41 && cp < 0x5B {
		return cp - 0x41
	}
	if cp >= 0x61 && cp < 0x7B {
		return cp - 0x61
	}
	return base
}

func digitToBasic(digit int32, flag bool) int32 {
	lt := int32(0)
	if digit < 26 {
		lt = 1
	}
	f := int32(0)
	if flag {
		f = 1
	}
	return digit + 22 + 75*lt - (f << 5)
}

func adapt(delta, numPoints int32, firstTime bool) int32 {
	if firstTime {
		delta /= damp
	} else {
		delta >>= 1
	}
	delta += delta / numPoints
	k := int32(0)
	for delta > baseMinusTMin*tMax>>1 {
		delta /= baseMinusTMin
		k += base
	}
	return k + (baseMinusTMin+1)*delta/(delta+skew)
}

// Decode converts a Punycode label (ASCII) to a Unicode string.
func Decode(input string) (string, error) {
	output := make([]int32, 0, len(input))
	inputLen := int32(len(input))
	i := int32(0)
	n := initialN
	bias := initialBias

	basic := int32(strings.LastIndex(input, string(delim)))
	if basic < 0 {
		basic = 0
	}

	for j := int32(0); j < basic; j++ {
		if input[j] >= 0x80 {
			return "", ErrNotBasic
		}
		output = append(output, int32(input[j]))
	}

	index := basic
	if basic > 0 {
		index = basic + 1
	}

	for index < inputLen {
		oldi := i
		w := int32(1)
		for k := base; ; k += base {
			if index >= inputLen {
				return "", ErrInvalidInput
			}
			digit := basicToDigit(int32(input[index]))
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

		out := int32(len(output)) + 1
		bias = adapt(i-oldi, out, oldi == 0)

		if i/out > maxInt-n {
			return "", ErrOverflow
		}
		n += i / out
		i %= out

		// Insert n at position i (equivalent to JS splice)
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = n
		i++
	}

	runes := make([]rune, len(output))
	for idx, v := range output {
		runes[idx] = rune(v)
	}
	return string(runes), nil
}

// Encode converts a Unicode label to Punycode (ASCII).
func Encode(input string) (string, error) {
	codePoints := make([]int32, 0, len([]rune(input)))
	for _, r := range input {
		codePoints = append(codePoints, int32(r))
	}

	inputLen := int32(len(codePoints))
	n := initialN
	delta := int32(0)
	bias := initialBias

	out := make([]byte, 0, len(input)+1)

	for _, cp := range codePoints {
		if cp < 0x80 {
			out = append(out, byte(cp))
		}
	}

	basicLen := int32(len(out))
	handled := basicLen

	if basicLen > 0 {
		out = append(out, delim)
	}

	for handled < inputLen {
		m := maxInt
		for _, cp := range codePoints {
			if cp >= n && cp < m {
				m = cp
			}
		}

		handledPlus1 := handled + 1
		if m-n > (maxInt-delta)/handledPlus1 {
			return "", ErrOverflow
		}
		delta += (m - n) * handledPlus1
		n = m

		for _, cp := range codePoints {
			if cp < n {
				if delta >= maxInt {
					return "", ErrOverflow
				}
				delta++
			} else if cp == n {
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
					out = append(out, byte(digitToBasic(t+qMinusT%baseMinusT, false)))
					q = qMinusT / baseMinusT
				}
				out = append(out, byte(digitToBasic(q, false)))
				bias = adapt(delta, handledPlus1, handled == basicLen)
				delta = 0
				handled++
			}
		}
		delta++
		n++
	}
	return string(out), nil
}

func mapDomain(input string, fn func(string) (string, error)) (string, error) {
	parts := strings.SplitN(input, "@", 2)
	prefix := ""
	domain := input
	if len(parts) == 2 {
		prefix = parts[0] + "@"
		domain = parts[1]
	}

	// Normalize IDNA2003 label separators to U+002E
	domain = strings.Map(func(r rune) rune {
		switch r {
		case '\u002E', '\u3002', '\uFF0E', '\uFF61':
			return '.'
		}
		return r
	}, domain)

	labels := strings.Split(domain, ".")
	for i, label := range labels {
		processed, err := fn(label)
		if err != nil {
			return "", err
		}
		labels[i] = processed
	}
	return prefix + strings.Join(labels, "."), nil
}

// ToUnicode converts a Punycode domain name or email address to Unicode.
func ToUnicode(input string) (string, error) {
	return mapDomain(input, func(label string) (string, error) {
		if strings.HasPrefix(label, "xn--") {
			return Decode(strings.ToLower(label[4:]))
		}
		return label, nil
	})
}

// ToASCII converts a Unicode domain name or email address to Punycode.
func ToASCII(input string) (string, error) {
	return mapDomain(input, func(label string) (string, error) {
		for _, r := range label {
			if r > 0x7F {
				encoded, err := Encode(label)
				if err != nil {
					return "", err
				}
				return "xn--" + encoded, nil
			}
		}
		return label, nil
	})
}
