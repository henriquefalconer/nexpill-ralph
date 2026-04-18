package punycode

import (
	"reflect"
	"testing"
)

// TestUCS2DecodeUTF16 verifies UCS2DecodeUTF16 against all seven fixtures
// (tests/tests.js:245-271, specs/punycode-ucs2-decode.md:13-21).
func TestUCS2DecodeUTF16(t *testing.T) {
	for _, fx := range ucs2Fixtures {
		fx := fx
		t.Run(fx.description, func(t *testing.T) {
			got := UCS2DecodeUTF16(fx.encoded)
			if !reflect.DeepEqual(got, fx.decoded) {
				t.Errorf("UCS2DecodeUTF16(%v) = %v, want %v", fx.encoded, got, fx.decoded)
			}
		})
	}
}

// TestUCS2EncodeUTF16 verifies UCS2EncodeUTF16 against all seven fixtures
// (tests/tests.js:273-288, specs/punycode-ucs2-encode.md:13-27).
func TestUCS2EncodeUTF16(t *testing.T) {
	for _, fx := range ucs2Fixtures {
		fx := fx
		t.Run(fx.description, func(t *testing.T) {
			got := UCS2EncodeUTF16(fx.decoded)
			if !reflect.DeepEqual(got, fx.encoded) {
				t.Errorf("UCS2EncodeUTF16(%v) = %v, want %v", fx.decoded, got, fx.encoded)
			}
		})
	}
}

// TestUCS2EncodeImmutability asserts that UCS2EncodeUTF16 does not mutate its input
// (tests/tests.js:282-287, specs/punycode-ucs2-encode.md:27).
func TestUCS2EncodeImmutability(t *testing.T) {
	input := []rune{0xD800}
	before := []rune{0xD800}
	UCS2EncodeUTF16(input)
	if !reflect.DeepEqual(input, before) {
		t.Errorf("UCS2EncodeUTF16 mutated input: got %v, want %v", input, before)
	}
}

// TestUCS2DecodeNormalUTF8 verifies that UCS2Decode handles well-formed UTF-8 strings
// including supplementary-plane code points (tests/tests.js:141-144 first fixture).
func TestUCS2DecodeNormalUTF8(t *testing.T) {
	input := "\U0001F355\U0001D400\U0001D306\U0001D356"
	want := []rune{0x1F355, 0x1D400, 0x1D306, 0x1D356}
	got := UCS2Decode(input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UCS2Decode(%q) = %v, want %v", input, got, want)
	}
}
