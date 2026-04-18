package punycode

import "strings"

// separatorReplacer normalises IDNA2003 full-stop variants to U+002E.
// punycode.js:19, 82 — regexSeparators: /[\x2E\u3002\uFF0E\uFF61]/g
var separatorReplacer = strings.NewReplacer(
	"\u3002", ".",
	"\uFF0E", ".",
	"\uFF61", ".",
)

// mapDomain splits an optional email prefix, normalises label separators,
// then applies callback to every dot-delimited label.
// Mirrors punycode.js:62-86; specs/src-punycode.md:74-80.
func mapDomain(domain string, callback func(string) (string, error)) (string, error) {
	// Preserve the local part of an email address. punycode.js:65-71
	prefix := ""
	if idx := strings.LastIndex(domain, "@"); idx >= 0 {
		prefix = domain[:idx+1]
		domain = domain[idx+1:]
	}

	// Normalise separators. punycode.js:73-74
	domain = separatorReplacer.Replace(domain)

	labels := strings.Split(domain, ".")
	out := make([]string, len(labels))
	for i, label := range labels {
		result, err := callback(label)
		if err != nil {
			return "", err
		}
		out[i] = result
	}

	return prefix + strings.Join(out, "."), nil
}
