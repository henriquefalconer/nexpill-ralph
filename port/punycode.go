package punycode

import "strings"

// ToUnicode converts a Punycode domain name or email address to Unicode.
// Labels prefixed with "xn--" are decoded; others pass through unchanged.
// Errors from Decode are swallowed — the label is returned as-is on failure,
// matching punycode.js:391-393 where decode throws but toUnicode catches.
// Mirrors punycode.js:389-395; specs/punycode-to-unicode.md.
func ToUnicode(input string) string {
	result, _ := mapDomain(input, func(label string) (string, error) {
		if strings.HasPrefix(label, "xn--") {
			decoded, err := Decode(strings.ToLower(label[4:]))
			if err != nil {
				return label, nil //nolint — swallow per JS semantics
			}
			return decoded, nil
		}
		return label, nil
	})
	return result
}

// ToASCII converts a Unicode domain name or email address to Punycode/ASCII.
// Labels containing non-ASCII bytes are encoded as "xn--" + Encode(label).
// Labels that are already ASCII pass through unchanged.
// Mirrors punycode.js:408-414; specs/punycode-to-ascii.md.
func ToASCII(input string) (string, error) {
	return mapDomain(input, func(label string) (string, error) {
		// Mirror regexNonASCII /[^\0-\x7F]/ — DEL (0x7F) is ASCII. punycode.js:18
		hasNonASCII := false
		for i := 0; i < len(label); i++ {
			if label[i] > 0x7F {
				hasNonASCII = true
				break
			}
		}
		if hasNonASCII {
			encoded, err := Encode(label)
			if err != nil {
				return "", err
			}
			return "xn--" + encoded, nil
		}
		return label, nil
	})
}
