package punycode

import "testing"

func TestEncode(t *testing.T) {
	// tests/tests.js:313-319
	for _, fx := range stringsFixtures {
		desc := fx.description
		if desc == "" {
			desc = fx.decoded
		}
		t.Run(desc, func(t *testing.T) {
			got, err := Encode(fx.decoded)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", fx.decoded, err)
			}
			if got != fx.encoded {
				t.Errorf("Encode(%q)\n  got  %q\n  want %q", fx.decoded, got, fx.encoded)
			}
		})
	}
}

// TestEncodeDecodeRoundTrip asserts Decode(Encode(decoded)) == decoded for every strings fixture.
// This is the belts-and-braces round-trip check from the todo.
func TestEncodeDecodeRoundTrip(t *testing.T) {
	for _, fx := range stringsFixtures {
		desc := fx.description
		if desc == "" {
			desc = fx.decoded
		}
		t.Run(desc, func(t *testing.T) {
			enc, err := Encode(fx.decoded)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", fx.decoded, err)
			}
			got, err := Decode(enc)
			if err != nil {
				t.Fatalf("Decode(Encode(%q)) error: %v", fx.decoded, err)
			}
			if got != fx.decoded {
				t.Errorf("round-trip(%q)\n  got  %q", fx.decoded, got)
			}
		})
	}
}
