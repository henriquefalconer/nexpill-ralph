package punycode

import "testing"

// domainVector mirrors one entry in testData.domains from tests/tests.js:176-220.
type domainVector struct {
	description  string
	decoded      string // Unicode domain or email
	encoded      string // ACE/IDNA encoded domain
	skipToASCII  bool   // ToASCII cannot reproduce this from the decoded form in Go
}

// domainVectors holds all 10 vectors from tests/tests.js:176-220.
// Vector 6 (lone surrogate U+D400): the decoded form is represented using
// WTF-8 raw bytes (\xed\x90\x80 = WTF-8 for U+D400) because Go string
// literals reject surrogate code points.  ToASCII cannot round-trip this
// vector — Go's []rune(s) maps the WTF-8 surrogate bytes to U+FFFD — so
// that direction is skipped with skipToASCII=true.
var domainVectors = []domainVector{
	{
		decoded: "ma\u00F1ana.com",
		encoded: "xn--maana-pta.com",
	},
	{
		decoded: "example.com.",
		encoded: "example.com.",
	},
	{
		decoded: "b\u00FCcher.com",
		encoded: "xn--bcher-kva.com",
	},
	{
		decoded: "caf\u00E9.com",
		encoded: "xn--caf-dma.com",
	},
	{
		decoded: "\u2603-\u2318.com",
		encoded: "xn----dqo34k.com",
	},
	{
		// Lone high surrogate U+D400 — represented as WTF-8 raw bytes.
		// tests/tests.js:197-200
		description: "lone surrogate (WTF-8 U+D400)",
		decoded:     "\xed\x90\x80\xe2\x98\x83-\xe2\x8c\x98.com",
		encoded:     "xn----dqo34kn65z.com",
		skipToASCII: true,
	},
	{
		// Pile of poo emoji U+1F4A9 — in JS stored as surrogate pair \uD83D\uDCA9.
		// tests/tests.js:201-205
		description: "Emoji",
		decoded:     "\U0001F4A9.la",
		encoded:     "xn--ls8h.la",
	},
	{
		// Non-printable ASCII; all bytes < 0x80 so no encoding occurs.
		// tests/tests.js:206-210
		description: "Non-printable ASCII",
		decoded:     "\x00\x01\x02foo.bar",
		encoded:     "\x00\x01\x02foo.bar",
	},
	{
		// Email address: Cyrillic local part preserved verbatim.
		// tests/tests.js:211-215
		description: "Email address",
		decoded:     "\u0434\u0436\u0443\u043C\u043B\u0430@\u0434\u0436p\u0443\u043C\u043B\u0430\u0442\u0435\u0441\u0442.b\u0440\u0444a",
		encoded:     "\u0434\u0436\u0443\u043C\u043B\u0430@xn--p-8sbkgc5ag7bhce.xn--ba-lmcq",
	},
	{
		// DEL (U+007F) is ≤ 0x7F so regexNonASCII does not match.
		// tests/tests.js:216-219
		decoded: "foo\x7F.example",
		encoded: "foo\x7F.example",
	},
}

// separatorVectors holds all 4 vectors from tests/tests.js:221-242.
// Each exercises a different RFC 3490 label separator that mapDomain normalises
// to U+002E before splitting.
var separatorVectors = []struct {
	description string
	decoded     string
	encoded     string
}{
	{
		description: "Using U+002E as separator",
		decoded:     "ma\u00F1ana.com",
		encoded:     "xn--maana-pta.com",
	},
	{
		description: "Using U+3002 as separator",
		decoded:     "ma\u00F1ana\u3002com",
		encoded:     "xn--maana-pta.com",
	},
	{
		description: "Using U+FF0E as separator",
		decoded:     "ma\u00F1ana\uFF0Ecom",
		encoded:     "xn--maana-pta.com",
	},
	{
		description: "Using U+FF61 as separator",
		decoded:     "ma\u00F1ana\uFF61com",
		encoded:     "xn--maana-pta.com",
	},
}

func TestToUnicode(t *testing.T) {
	// Domain conversion: ToUnicode(encoded) === decoded (10 vectors).
	for i, v := range domainVectors {
		name := v.description
		if name == "" {
			name = v.encoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := ToUnicode(v.encoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", v.encoded, err)
			}
			if got != v.decoded {
				t.Errorf("ToUnicode(%q) [vector %d]\n got: %q\nwant: %q", v.encoded, i, got, v.decoded)
			}
		})
	}

	// Pass-through: bare Punycode tokens without xn-- prefix are returned unchanged.
	// Asserts both ToUnicode(encoded)===encoded and ToUnicode(decoded)===decoded.
	// Source: tests/tests.js:332-343 (iterates testData.strings, 24 vectors).
	for _, v := range decodeVectors {
		name := v.description
		if name == "" {
			name = v.encoded
		}
		t.Run("passthrough/encoded/"+name, func(t *testing.T) {
			got, err := ToUnicode(v.encoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", v.encoded, err)
			}
			if got != v.encoded {
				t.Errorf("ToUnicode(%q) = %q, want pass-through %q", v.encoded, got, v.encoded)
			}
		})
		t.Run("passthrough/decoded/"+name, func(t *testing.T) {
			got, err := ToUnicode(v.decoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", v.decoded, err)
			}
			if got != v.decoded {
				t.Errorf("ToUnicode(%q) = %q, want pass-through %q", v.decoded, got, v.decoded)
			}
		})
	}
}

func TestToASCII(t *testing.T) {
	// Domain conversion: ToASCII(decoded) === encoded (10 vectors, vector 6 skipped).
	for i, v := range domainVectors {
		if v.skipToASCII {
			continue
		}
		name := v.description
		if name == "" {
			name = v.decoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := ToASCII(v.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", v.decoded, err)
			}
			if got != v.encoded {
				t.Errorf("ToASCII(%q) [vector %d]\n got: %q\nwant: %q", v.decoded, i, got, v.encoded)
			}
		})
	}

	// Pass-through: already-ASCII strings (encoded forms) are returned unchanged.
	// Source: tests/tests.js:355-362 (iterates testData.strings, 24 vectors).
	for _, v := range decodeVectors {
		name := v.description
		if name == "" {
			name = v.encoded
		}
		t.Run("passthrough/"+name, func(t *testing.T) {
			got, err := ToASCII(v.encoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", v.encoded, err)
			}
			if got != v.encoded {
				t.Errorf("ToASCII(%q) = %q, want pass-through %q", v.encoded, got, v.encoded)
			}
		})
	}

	// IDNA2003 separator normalization: all four dot variants normalize to U+002E.
	// Source: tests/tests.js:363-370 (4 vectors).
	for _, v := range separatorVectors {
		t.Run(v.description, func(t *testing.T) {
			got, err := ToASCII(v.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", v.decoded, err)
			}
			if got != v.encoded {
				t.Errorf("ToASCII(%q) = %q, want %q", v.decoded, got, v.encoded)
			}
		})
	}
}
