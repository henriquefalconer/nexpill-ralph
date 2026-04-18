package punycode

import "testing"

// TestToUnicodeDescribe mirrors describe('punycode.toUnicode') at tests/tests.js:322-344.
func TestToUnicodeDescribe(t *testing.T) {
	// Domain fixtures: tests/tests.js:323-330
	for _, f := range domainFixtures {
		f := f
		desc := f.description
		if desc == "" {
			desc = f.encoded
		}
		t.Run(desc, func(t *testing.T) {
			got := ToUnicode(f.encoded)
			if got != f.decoded {
				t.Fatalf("ToUnicode(%q) = %q, want %q", f.encoded, got, f.decoded)
			}
		})
	}

	// Identity loop: strings fixtures don't start with xn-- so they pass through.
	// tests/tests.js:332-343
	for _, f := range stringsFixtures {
		f := f
		t.Run("no-convert/"+f.encoded, func(t *testing.T) {
			if got := ToUnicode(f.encoded); got != f.encoded {
				t.Fatalf("ToUnicode(%q) = %q, want identity", f.encoded, got)
			}
			if got := ToUnicode(f.decoded); got != f.decoded {
				t.Fatalf("ToUnicode(%q) = %q, want identity", f.decoded, got)
			}
		})
	}
}

// TestToASCIIDescribe mirrors describe('punycode.toASCII') at tests/tests.js:345-371.
func TestToASCIIDescribe(t *testing.T) {
	// Domain fixtures: tests/tests.js:347-353
	for _, f := range domainFixtures {
		f := f
		desc := f.description
		if desc == "" {
			desc = f.decoded
		}
		t.Run(desc, func(t *testing.T) {
			got, err := ToASCII(f.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", f.decoded, err)
			}
			if got != f.encoded {
				t.Fatalf("ToASCII(%q) = %q, want %q", f.decoded, got, f.encoded)
			}
		})
	}

	// Identity loop: strings fixtures are pure ASCII. tests/tests.js:355-361
	for _, f := range stringsFixtures {
		f := f
		t.Run("no-convert/"+f.encoded, func(t *testing.T) {
			got, err := ToASCII(f.encoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", f.encoded, err)
			}
			if got != f.encoded {
				t.Fatalf("ToASCII(%q) = %q, want identity", f.encoded, got)
			}
		})
	}

	// Separator normalisation: tests/tests.js:363-370
	for _, f := range separatorFixtures {
		f := f
		t.Run(f.description, func(t *testing.T) {
			got, err := ToASCII(f.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", f.decoded, err)
			}
			if got != f.encoded {
				t.Fatalf("ToASCII(%q) = %q, want %q", f.decoded, got, f.encoded)
			}
		})
	}
}
