package punycode

import "testing"

func TestToUnicode(t *testing.T) {
	// Domain conversion: ToUnicode(encoded) === decoded (10 vectors).
	for i, v := range testVectors.Domains {
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
	for _, v := range testVectors.Strings {
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
	for i, v := range testVectors.Domains {
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
	for _, v := range testVectors.Strings {
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
	for _, v := range testVectors.Separators {
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
