// Package punycode implements the Punycode algorithm (RFC 3492).
// It is a faithful Go port of punycode.js v2.3.1 by Mathias Bynens.
package punycode

import (
	"errors"
	"strings"
	"unicode/utf16"
)

// version matches the ported JS library.
const version = "2.3.1"

// Bootstring parameters (RFC 3492 §5).
const (
	base          = 36
	tMin          = 1
	tMax          = 26
	skew          = 38
	damp          = 700
	initialBias   = 72
	initialN      = 128
	delimiter     = '-'
	maxInt        = 2147483647 // 0x7FFFFFFF, 2^31-1
	baseMinusTMin = base - tMin
)

// Exported errors; message strings match punycode.js exactly.
var (
	ErrOverflow     = errors.New("Overflow: input needs wider integers to process")
	ErrNotBasic     = errors.New("Illegal input >= 0x80 (not a basic code point)")
	ErrInvalidInput = errors.New("Invalid input")
)

// hasPunycodePrefix reports whether s starts with the ACE prefix "xn--".
func hasPunycodePrefix(s string) bool {
	return strings.HasPrefix(s, "xn--")
}

// hasNonASCII reports whether s contains any code point > U+007F.
func hasNonASCII(s string) bool {
	for _, r := range s {
		if r > 0x7F {
			return true
		}
	}
	return false
}

// normalizeSeparators replaces RFC 3490 label separators with '.'.
func normalizeSeparators(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\x2E', '\u3002', '\uFF0E', '\uFF61':
			return '.'
		}
		return r
	}, s)
}

// basicToDigit converts an ASCII code point to its Bootstring digit value
// (0–35), or base (36) if not a valid digit character.
func basicToDigit(cp int) int {
	switch {
	case cp >= 0x30 && cp < 0x3A:
		return 26 + (cp - 0x30)
	case cp >= 0x41 && cp < 0x5B:
		return cp - 0x41
	case cp >= 0x61 && cp < 0x7B:
		return cp - 0x61
	}
	return base
}

// digitToBasic converts a Bootstring digit value to its ASCII code point.
// flag=0 gives lowercase; flag!=0 gives uppercase.
func digitToBasic(digit, flag int) int {
	result := digit + 22
	if digit < 26 {
		result += 75
	}
	if flag != 0 {
		result -= 32
	}
	return result
}

// adapt implements the RFC 3492 §3.4 bias adaptation algorithm.
func adapt(delta, numPoints int, firstTime bool) int {
	if firstTime {
		delta /= damp
	} else {
		delta >>= 1
	}
	delta += delta / numPoints
	k := 0
	for delta > baseMinusTMin*tMax>>1 {
		delta /= baseMinusTMin
		k += base
	}
	return k + (baseMinusTMin+1)*delta/(delta+skew)
}

// UCS2Decode converts a Go string to a slice of Unicode code points using
// UTF-16 semantics: surrogate pairs are combined; lone surrogates are emitted
// as their raw code-unit values.
func UCS2Decode(s string) []int32 {
	units := utf16.Encode([]rune(s))
	output := make([]int32, 0, len(units))
	i := 0
	for i < len(units) {
		value := int32(units[i])
		i++
		if value >= 0xD800 && value <= 0xDBFF && i < len(units) {
			extra := int32(units[i])
			if (extra & 0xFC00) == 0xDC00 {
				output = append(output, ((value&0x3FF)<<10)+(extra&0x3FF)+0x10000)
				i++
				continue
			}
		}
		output = append(output, value)
	}
	return output
}

// UCS2Encode converts a slice of Unicode code points to a Go string using
// UTF-16 semantics. The input slice is never mutated.
func UCS2Encode(codePoints []int32) string {
	units := make([]uint16, 0, len(codePoints))
	for _, cp := range codePoints {
		if cp >= 0x10000 {
			cp -= 0x10000
			units = append(units, uint16(0xD800+(cp>>10)))
			units = append(units, uint16(0xDC00+(cp&0x3FF)))
		} else {
			units = append(units, uint16(cp))
		}
	}
	return string(utf16.Decode(units))
}

// decodeInternal is the inner decoder; panics on invalid input.
func decodeInternal(input string) string {
	var output []int32
	inputLength := len(input)
	i := 0
	n := initialN
	bias := initialBias

	basic := strings.LastIndex(input, string(delimiter))
	if basic < 0 {
		basic = 0
	}
	for j := 0; j < basic; j++ {
		if input[j] >= 0x80 {
			panic(ErrNotBasic)
		}
		output = append(output, int32(input[j]))
	}

	index := 0
	if basic > 0 {
		index = basic + 1
	}

	for index < inputLength {
		oldi := i
		w := 1
		for k := base; ; k += base {
			if index >= inputLength {
				panic(ErrInvalidInput)
			}
			digit := basicToDigit(int(input[index]))
			index++
			if digit >= base {
				panic(ErrInvalidInput)
			}
			if digit > (maxInt-i)/w {
				panic(ErrOverflow)
			}
			i += digit * w
			t := tMin
			if k >= bias+tMax {
				t = tMax
			} else if k > bias {
				t = k - bias
			}
			if digit < t {
				break
			}
			baseMinusT := base - t
			if w > maxInt/baseMinusT {
				panic(ErrOverflow)
			}
			w *= baseMinusT
		}

		out := len(output) + 1
		bias = adapt(i-oldi, out, oldi == 0)

		if i/out > maxInt-n {
			panic(ErrOverflow)
		}
		n += i / out
		i %= out

		// Insert n at position i.
		output = append(output, 0)
		copy(output[i+1:], output[i:])
		output[i] = int32(n)
		i++
	}

	return UCS2Encode(output)
}

// encodeInternal is the inner encoder; panics on overflow.
func encodeInternal(input string) string {
	codePoints := UCS2Decode(input)
	inputLength := len(codePoints)

	var out strings.Builder
	n := initialN
	delta := 0
	bias := initialBias

	for _, cp := range codePoints {
		if cp < 0x80 {
			out.WriteByte(byte(cp))
		}
	}

	basicLength := out.Len()
	handledCPCount := basicLength

	if basicLength > 0 {
		out.WriteByte(delimiter)
	}

	for handledCPCount < inputLength {
		m := maxInt
		for _, cp := range codePoints {
			if int(cp) >= n && int(cp) < m {
				m = int(cp)
			}
		}

		handledCPCountPlusOne := handledCPCount + 1
		if m-n > (maxInt-delta)/handledCPCountPlusOne {
			panic(ErrOverflow)
		}
		delta += (m - n) * handledCPCountPlusOne
		n = m

		for _, cp := range codePoints {
			if int(cp) < n {
				delta++
				if delta > maxInt {
					panic(ErrOverflow)
				}
			}
			if int(cp) == n {
				q := delta
				for k := base; ; k += base {
					t := tMin
					if k >= bias+tMax {
						t = tMax
					} else if k > bias {
						t = k - bias
					}
					if q < t {
						break
					}
					qMinusT := q - t
					baseMinusT := base - t
					out.WriteByte(byte(digitToBasic(t+qMinusT%baseMinusT, 0)))
					q = qMinusT / baseMinusT
				}
				out.WriteByte(byte(digitToBasic(q, 0)))
				bias = adapt(delta, handledCPCountPlusOne, handledCPCount == basicLength)
				delta = 0
				handledCPCount++
			}
		}
		delta++
		n++
	}
	return out.String()
}

// Decode converts a Punycode string of ASCII-only symbols to a string of
// Unicode symbols.
func Decode(input string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()
	return decodeInternal(input), nil
}

// Encode converts a string of Unicode symbols to a Punycode string of
// ASCII-only symbols.
func Encode(input string) (result string, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			}
		}
	}()
	return encodeInternal(input), nil
}

// mapDomain applies fn to each label of a domain name (or email address).
// The local part of an email address (before the first '@') is preserved.
func mapDomain(s string, fn func(string) (string, error)) (string, error) {
	atIdx := strings.IndexByte(s, '@')
	localPart := ""
	domain := s
	if atIdx >= 0 {
		localPart = s[:atIdx+1]
		domain = s[atIdx+1:]
	}
	domain = normalizeSeparators(domain)
	labels := strings.Split(domain, ".")
	for i, label := range labels {
		result, err := fn(label)
		if err != nil {
			return "", err
		}
		labels[i] = result
	}
	return localPart + strings.Join(labels, "."), nil
}

// ToUnicode converts a Punycoded domain name or email address to Unicode.
// Only labels beginning with "xn--" are decoded; others pass through.
func ToUnicode(input string) (string, error) {
	return mapDomain(input, func(label string) (string, error) {
		if hasPunycodePrefix(label) {
			return Decode(strings.ToLower(label[4:]))
		}
		return label, nil
	})
}

// ToASCII converts a Unicode domain name or email address to Punycode.
// Only labels containing non-ASCII characters are encoded; others pass through.
func ToASCII(input string) (string, error) {
	return mapDomain(input, func(label string) (string, error) {
		if hasNonASCII(label) {
			encoded, err := Encode(label)
			if err != nil {
				return "", err
			}
			return "xn--" + encoded, nil
		}
		return label, nil
	})
}
