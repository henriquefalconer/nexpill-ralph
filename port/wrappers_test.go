package punycode

import "testing"

func TestToUnicode(t *testing.T) {
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

	// Identity: pure-ASCII encoded strings pass through unchanged. tests/tests.js:332-343
	for _, f := range stringsFixtures {
		f := f
		t.Run("identity/encoded/"+f.encoded, func(t *testing.T) {
			if got := ToUnicode(f.encoded); got != f.encoded {
				t.Fatalf("ToUnicode(%q) = %q, want identity", f.encoded, got)
			}
		})
		t.Run("identity/decoded/"+f.decoded, func(t *testing.T) {
			if got := ToUnicode(f.decoded); got != f.decoded {
				t.Fatalf("ToUnicode(%q) = %q, want identity", f.decoded, got)
			}
		})
	}
}

func TestToASCII(t *testing.T) {
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

	// Identity: pure-ASCII encoded strings need no encoding. tests/tests.js:355-361
	for _, f := range stringsFixtures {
		f := f
		t.Run("identity/"+f.encoded, func(t *testing.T) {
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

	// DEL (0x7F) is ASCII — no encoding applied. tests/tests.js:216-219
	t.Run("DEL passthrough", func(t *testing.T) {
		got, err := ToASCII("foo\x7f.example")
		if err != nil || got != "foo\x7f.example" {
			t.Fatalf("ToASCII DEL: got %q, err %v", got, err)
		}
	})

	// Trailing-dot preservation: tests/tests.js:181-184
	t.Run("trailing dot preserved", func(t *testing.T) {
		got, err := ToASCII("example.com.")
		if err != nil || got != "example.com." {
			t.Fatalf("ToASCII trailing dot: got %q, err %v", got, err)
		}
	})
}
