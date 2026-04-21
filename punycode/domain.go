package punycode

import "strings"

// mapDomain applies fn to each dot-separated label of domain (or the domain
// part of an email address), rejoining with ".".
// Mirrors punycode.js:62-86.
func mapDomain(domain string, fn func(string) (string, error)) (string, error) {
	parts := strings.SplitN(domain, "@", 2)
	prefix := ""
	if len(parts) > 1 {
		prefix = parts[0] + "@"
		domain = parts[1]
	}
	// Normalize RFC 3490 separators to U+002E before splitting.
	// U+002E is already ".", so omit it from the replacer.
	replacer := strings.NewReplacer("\u3002", ".", "\uFF0E", ".", "\uFF61", ".")
	domain = replacer.Replace(domain)
	labels := strings.Split(domain, ".")
	for i, label := range labels {
		transformed, err := fn(label)
		if err != nil {
			return "", err
		}
		labels[i] = transformed
	}
	return prefix + strings.Join(labels, "."), nil
}

// ToASCII converts a Unicode domain name or email address to its Punycode
// (ACE) representation. Labels containing non-ASCII characters are encoded
// and prefixed with "xn--"; pure-ASCII labels pass through unchanged.
// Mirrors punycode.js:408-414.
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

// ToUnicode converts a Punycode domain name or email address to Unicode.
// Labels prefixed with "xn--" are decoded; all other labels pass through.
// Mirrors punycode.js:389-395.
func ToUnicode(input string) (string, error) {
	return mapDomain(input, func(label string) (string, error) {
		if strings.HasPrefix(label, "xn--") {
			decoded, err := Decode(strings.ToLower(label[4:]))
			if err != nil {
				return "", err
			}
			return decoded, nil
		}
		return label, nil
	})
}
