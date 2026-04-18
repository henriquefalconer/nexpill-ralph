package punycode

import (
	"reflect"
	"testing"
)

// ucs2Fixture mirrors one record from testData.ucs2 (tests/tests.js:137-175).
// Encoded is []uint16 (not string) to carry lone surrogates faithfully.
type ucs2Fixture struct {
	description string
	decoded     []rune
	encoded     []uint16
}

// ucs2Fixtures are the seven combinations fixtures from tests/tests.js:137-175.
var ucs2Fixtures = []ucs2Fixture{
	{
		description: "Consecutive astral symbols",
		decoded:     []rune{127829, 119808, 119558, 119638},
		// '\uD83C\uDF55\uD835\uDC00\uD834\uDF06\uD834\uDF56'
		encoded: []uint16{0xD83C, 0xDF55, 0xD835, 0xDC00, 0xD834, 0xDF06, 0xD834, 0xDF56},
	},
	{
		description: "U+D800 (high surrogate) followed by non-surrogates",
		decoded:     []rune{0xD800, 97, 98},
		// '\uD800ab'
		encoded: []uint16{0xD800, 0x61, 0x62},
	},
	{
		description: "U+DC00 (low surrogate) followed by non-surrogates",
		decoded:     []rune{0xDC00, 97, 98},
		// '\uDC00ab'
		encoded: []uint16{0xDC00, 0x61, 0x62},
	},
	{
		description: "High surrogate followed by another high surrogate",
		decoded:     []rune{0xD800, 0xD800},
		// '\uD800\uD800'
		encoded: []uint16{0xD800, 0xD800},
	},
	{
		description: "Unmatched high surrogate, followed by a surrogate pair, followed by an unmatched high surrogate",
		decoded:     []rune{0xD800, 0x1D306, 0xD800},
		// '\uD800\uD834\uDF06\uD800'
		encoded: []uint16{0xD800, 0xD834, 0xDF06, 0xD800},
	},
	{
		description: "Low surrogate followed by another low surrogate",
		decoded:     []rune{0xDC00, 0xDC00},
		// '\uDC00\uDC00'
		encoded: []uint16{0xDC00, 0xDC00},
	},
	{
		description: "Unmatched low surrogate, followed by a surrogate pair, followed by an unmatched low surrogate",
		decoded:     []rune{0xDC00, 0x1D306, 0xDC00},
		// '\uDC00\uD834\uDF06\uDC00'
		encoded: []uint16{0xDC00, 0xD834, 0xDF06, 0xDC00},
	},
}

// TestUCS2DecodeUTF16 verifies UCS2DecodeUTF16 against all seven fixtures
// (tests/tests.js:245-271, punycode-ucs2-decode.md:13-21).
func TestUCS2DecodeUTF16(t *testing.T) {
	for _, fx := range ucs2Fixtures {
		t.Run(fx.description, func(t *testing.T) {
			got := UCS2DecodeUTF16(fx.encoded)
			if !reflect.DeepEqual(got, fx.decoded) {
				t.Errorf("UCS2DecodeUTF16(%v) = %v, want %v", fx.encoded, got, fx.decoded)
			}
		})
	}
}

// TestUCS2EncodeUTF16 verifies UCS2EncodeUTF16 against all seven fixtures
// (tests/tests.js:273-288, punycode-ucs2-encode.md:13-27).
func TestUCS2EncodeUTF16(t *testing.T) {
	for _, fx := range ucs2Fixtures {
		t.Run(fx.description, func(t *testing.T) {
			got := UCS2EncodeUTF16(fx.decoded)
			if !reflect.DeepEqual(got, fx.encoded) {
				t.Errorf("UCS2EncodeUTF16(%v) = %v, want %v", fx.decoded, got, fx.encoded)
			}
		})
	}
}

// TestUCS2EncodeImmutability asserts that UCS2EncodeUTF16 does not mutate its input
// (tests/tests.js:282-287, punycode-ucs2-encode.md:27).
func TestUCS2EncodeImmutability(t *testing.T) {
	input := []rune{0xD800}
	before := []rune{0xD800}
	UCS2EncodeUTF16(input)
	if !reflect.DeepEqual(input, before) {
		t.Errorf("UCS2EncodeUTF16 mutated input: got %v, want %v", input, before)
	}
}

// TestUCS2DecodeNormalUTF8 verifies that UCS2Decode handles well-formed UTF-8 strings,
// including supplementary-plane code points (tests/tests.js:141-144 first fixture).
func TestUCS2DecodeNormalUTF8(t *testing.T) {
	// The first fixture's code points expressed as a UTF-8 string: 🍕𝐀𝄆𝄖
	input := "\U0001F355\U0001D400\U0001D306\U0001D356"
	want := []rune{0x1F355, 0x1D400, 0x1D306, 0x1D356}
	got := UCS2Decode(input)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("UCS2Decode(%q) = %v, want %v", input, got, want)
	}
}
