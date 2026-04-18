package punycode_test

import (
	"testing"

	"github.com/mathiasbynens/punycode"
)

func TestUCS2Decode(t *testing.T) {
	for _, tc := range ucs2Data {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got := punycode.UCS2Decode(tc.encoded)
			if len(got) != len(tc.decoded) {
				t.Fatalf("UCS2Decode(%q): got len %d, want %d", tc.encoded, len(got), len(tc.decoded))
			}
			for i := range got {
				if got[i] != tc.decoded[i] {
					t.Errorf("UCS2Decode(%q)[%d] = %d, want %d", tc.encoded, i, got[i], tc.decoded[i])
				}
			}
		})
	}
}

func TestUCS2Encode(t *testing.T) {
	for _, tc := range ucs2Data {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got := punycode.UCS2Encode(tc.decoded)
			if got != tc.encoded {
				t.Errorf("UCS2Encode(%v) = %q, want %q", tc.decoded, got, tc.encoded)
			}
		})
	}

	t.Run("does not mutate argument slice", func(t *testing.T) {
		codePoints := []int32{0x61, 0x62, 0x63}
		orig := []int32{0x61, 0x62, 0x63}
		result := punycode.UCS2Encode(codePoints)
		if result != "abc" {
			t.Errorf("UCS2Encode([0x61,0x62,0x63]) = %q, want %q", result, "abc")
		}
		for i := range codePoints {
			if codePoints[i] != orig[i] {
				t.Errorf("UCS2Encode mutated input at index %d: got %d, want %d", i, codePoints[i], orig[i])
			}
		}
	})
}

func TestDecode(t *testing.T) {
	for _, tc := range stringsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.encoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := punycode.Decode(tc.encoded)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", tc.encoded, err)
			}
			if got != tc.decoded {
				t.Errorf("Decode(%q) = %q, want %q", tc.encoded, got, tc.decoded)
			}
		})
	}

	t.Run("handles uppercase letters (case-insensitive)", func(t *testing.T) {
		got, err := punycode.Decode("ZZZ")
		if err != nil {
			t.Fatalf("Decode(\"ZZZ\") error: %v", err)
		}
		want := "\u7BA5"
		if got != want {
			t.Errorf("Decode(\"ZZZ\") = %q, want %q", got, want)
		}
	})

	t.Run("ErrNotBasic on non-basic prefix", func(t *testing.T) {
		_, err := punycode.Decode("\x81-")
		if err == nil {
			t.Error("Decode(\"\\x81-\") should return error, got nil")
		}
	})

	t.Run("error on invalid non-basic input", func(t *testing.T) {
		_, err := punycode.Decode("\x81")
		if err == nil {
			t.Error("Decode(\"\\x81\") should return error, got nil")
		}
	})

	t.Run("ErrInvalidInput on invalid digit", func(t *testing.T) {
		_, err := punycode.Decode("ls8h=")
		if err == nil {
			t.Error("Decode(\"ls8h=\") should return error, got nil")
		}
	})
}

func TestEncode(t *testing.T) {
	for _, tc := range stringsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.decoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := punycode.Encode(tc.decoded)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", tc.decoded, err)
			}
			if got != tc.encoded {
				t.Errorf("Encode(%q) = %q, want %q", tc.decoded, got, tc.encoded)
			}
		})
	}
}

func TestToUnicode(t *testing.T) {
	for _, tc := range domainsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.encoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := punycode.ToUnicode(tc.encoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", tc.encoded, err)
			}
			if got != tc.decoded {
				t.Errorf("ToUnicode(%q) = %q, want %q", tc.encoded, got, tc.decoded)
			}
		})
	}

	// Non-xn-- strings pass through unchanged.
	t.Run("does not convert strings not starting with xn--", func(t *testing.T) {
		for _, tc := range stringsData {
			got, err := punycode.ToUnicode(tc.encoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", tc.encoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToUnicode(%q) = %q, want %q (unchanged)", tc.encoded, got, tc.encoded)
			}
			got2, err := punycode.ToUnicode(tc.decoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", tc.decoded, err)
			}
			if got2 != tc.decoded {
				t.Errorf("ToUnicode(%q) = %q, want %q (unchanged)", tc.decoded, got2, tc.decoded)
			}
		}
	})
}

func TestToASCII(t *testing.T) {
	for _, tc := range domainsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.decoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := punycode.ToASCII(tc.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", tc.decoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToASCII(%q) = %q, want %q", tc.decoded, got, tc.encoded)
			}
		})
	}

	// Already-ASCII strings pass through unchanged.
	t.Run("does not convert already-ASCII strings", func(t *testing.T) {
		for _, tc := range stringsData {
			got, err := punycode.ToASCII(tc.encoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", tc.encoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToASCII(%q) = %q, want %q (unchanged)", tc.encoded, got, tc.encoded)
			}
		}
	})

	// Separator normalization.
	for _, tc := range separatorsData {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got, err := punycode.ToASCII(tc.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", tc.decoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToASCII(%q) = %q, want %q", tc.decoded, got, tc.encoded)
			}
		})
	}
}

func TestRoundTrip(t *testing.T) {
	for _, tc := range stringsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.decoded
		}
		t.Run(name, func(t *testing.T) {
			// Decode(Encode(decoded)) == decoded
			enc, err := punycode.Encode(tc.decoded)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", tc.decoded, err)
			}
			dec, err := punycode.Decode(enc)
			if err != nil {
				t.Fatalf("Decode(Encode(%q)) error: %v", tc.decoded, err)
			}
			if dec != tc.decoded {
				t.Errorf("Decode(Encode(%q)) = %q, want %q", tc.decoded, dec, tc.decoded)
			}

			// Encode(Decode(encoded)) == encoded
			dec2, err := punycode.Decode(tc.encoded)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", tc.encoded, err)
			}
			enc2, err := punycode.Encode(dec2)
			if err != nil {
				t.Fatalf("Encode(Decode(%q)) error: %v", tc.encoded, err)
			}
			if enc2 != tc.encoded {
				t.Errorf("Encode(Decode(%q)) = %q, want %q", tc.encoded, enc2, tc.encoded)
			}
		})
	}
}
