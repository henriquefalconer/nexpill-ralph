package punycode

import (
	"errors"
	"testing"
)

func TestDecode(t *testing.T) {
	// Strings fixtures: tests/tests.js:291-297
	for _, fx := range stringsFixtures {
		fx := fx
		desc := fx.description
		if desc == "" {
			desc = fx.encoded
		}
		t.Run(desc, func(t *testing.T) {
			got, err := Decode(fx.encoded)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", fx.encoded, err)
			}
			if got != fx.decoded {
				t.Errorf("Decode(%q) = %q, want %q", fx.encoded, got, fx.decoded)
			}
		})
	}

	// Case-insensitive decode: tests/tests.js:299-301
	t.Run("ZZZ case-insensitive", func(t *testing.T) {
		got, err := Decode("ZZZ")
		if err != nil {
			t.Fatalf("Decode(\"ZZZ\") error: %v", err)
		}
		if got != "\u7BA5" {
			t.Errorf("Decode(\"ZZZ\") = %q, want U+7BA5", got)
		}
	})

	// ErrNotBasic: non-basic byte before the delimiter. tests/tests.js:255-260
	t.Run("ErrNotBasic on \\x81-", func(t *testing.T) {
		_, err := Decode("\x81-")
		if !errors.Is(err, ErrNotBasic) {
			t.Errorf("Decode(\"\\x81-\") = %v, want ErrNotBasic", err)
		}
	})

	// ErrInvalidInput: non-basic byte with no delimiter triggers digit >= base guard.
	// tests/tests.js:261-270 — JS throws RangeError (any); in Go the specific error is ErrInvalidInput.
	t.Run("ErrInvalidInput on \\x81", func(t *testing.T) {
		_, err := Decode("\x81")
		if !errors.Is(err, ErrInvalidInput) {
			t.Fatalf("Decode(\"\\x81\") = %v, want ErrInvalidInput", err)
		}
	})

	// ErrInvalidInput: '=' is not a Bootstring digit. tests/tests.js:302-309
	t.Run("ErrInvalidInput on ls8h=", func(t *testing.T) {
		_, err := Decode("ls8h=")
		if !errors.Is(err, ErrInvalidInput) {
			t.Errorf("Decode(\"ls8h=\") = %v, want ErrInvalidInput", err)
		}
	})
}
