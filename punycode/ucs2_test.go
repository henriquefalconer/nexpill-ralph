package punycode

import "testing"

// ucs2Vector is a round-trip pair: the UTF-16 unit sequence and its
// decoded code-point slice. All 7 vectors come from
// specs/test-ucs2-decode.md and specs/test-ucs2-encode.md.
type ucs2Vector struct {
	desc       string
	units      []uint16
	codePoints []rune
}

var ucs2Vectors = []ucs2Vector{
	{
		"consecutive astral symbols",
		[]uint16{0xD83C, 0xDF55, 0xD835, 0xDC00, 0xD834, 0xDF06, 0xD834, 0xDF56},
		[]rune{127829, 119808, 119558, 119638},
	},
	{
		"high surrogate followed by non-surrogates",
		[]uint16{0xD800, 0x61, 0x62},
		[]rune{0xD800, 97, 98},
	},
	{
		"low surrogate followed by non-surrogates",
		[]uint16{0xDC00, 0x61, 0x62},
		[]rune{0xDC00, 97, 98},
	},
	{
		"high surrogate followed by another high surrogate",
		[]uint16{0xD800, 0xD800},
		[]rune{0xD800, 0xD800},
	},
	{
		"unmatched high, surrogate pair, unmatched high",
		[]uint16{0xD800, 0xD834, 0xDF06, 0xD800},
		[]rune{0xD800, 0x1D306, 0xD800},
	},
	{
		"low surrogate followed by another low surrogate",
		[]uint16{0xDC00, 0xDC00},
		[]rune{0xDC00, 0xDC00},
	},
	{
		"unmatched low, surrogate pair, unmatched low",
		[]uint16{0xDC00, 0xD834, 0xDF06, 0xDC00},
		[]rune{0xDC00, 0x1D306, 0xDC00},
	},
}

// WTF-8 byte sequences for the expected encode output.
// Surrogates are encoded as 3-byte WTF-8; astral code points as 4-byte UTF-8.
// Byte sequences pre-computed from the code-point values in ucs2Vectors.
var ucs2EncodeExpected = [][]byte{
	// vector 1: valid astral symbols → standard UTF-8
	[]byte(string([]rune{127829, 119808, 119558, 119638})),
	// vector 2: U+D800 'a' 'b' — WTF-8 for U+D800: ED A0 80
	{0xED, 0xA0, 0x80, 0x61, 0x62},
	// vector 3: U+DC00 'a' 'b' — WTF-8 for U+DC00: ED B0 80
	{0xED, 0xB0, 0x80, 0x61, 0x62},
	// vector 4: U+D800 U+D800
	{0xED, 0xA0, 0x80, 0xED, 0xA0, 0x80},
	// vector 5: U+D800 U+1D306 U+D800 — U+1D306 UTF-8: F0 9D 8C 86
	{0xED, 0xA0, 0x80, 0xF0, 0x9D, 0x8C, 0x86, 0xED, 0xA0, 0x80},
	// vector 6: U+DC00 U+DC00
	{0xED, 0xB0, 0x80, 0xED, 0xB0, 0x80},
	// vector 7: U+DC00 U+1D306 U+DC00
	{0xED, 0xB0, 0x80, 0xF0, 0x9D, 0x8C, 0x86, 0xED, 0xB0, 0x80},
}

func TestUCS2Decode(t *testing.T) {
	tests := []struct {
		input string
		want  []rune
	}{
		{"", []rune{}},
		{"abc", []rune{'a', 'b', 'c'}},
		{"日本語", []rune{0x65E5, 0x672C, 0x8A9E}},
		{"😀", []rune{0x1F600}},
	}
	for _, tc := range tests {
		got := UCS2Decode(tc.input)
		if len(got) != len(tc.want) {
			t.Errorf("UCS2Decode(%q): len=%d, want %d", tc.input, len(got), len(tc.want))
			continue
		}
		for i := range tc.want {
			if got[i] != tc.want[i] {
				t.Errorf("UCS2Decode(%q)[%d] = %U, want %U", tc.input, i, got[i], tc.want[i])
			}
		}
	}
}

func TestUCS2DecodeUnits(t *testing.T) {
	for _, v := range ucs2Vectors {
		got := UCS2DecodeUnits(v.units)
		if len(got) != len(v.codePoints) {
			t.Errorf("[%s]: len=%d, want %d", v.desc, len(got), len(v.codePoints))
			continue
		}
		for i, want := range v.codePoints {
			if got[i] != want {
				t.Errorf("[%s][%d] = %U, want %U", v.desc, i, got[i], want)
			}
		}
	}
}

func TestUCS2Encode(t *testing.T) {
	for i, v := range ucs2Vectors {
		got := UCS2Encode(v.codePoints)
		want := string(ucs2EncodeExpected[i])
		if got != want {
			t.Errorf("[%s]: got bytes %x, want %x", v.desc, []byte(got), ucs2EncodeExpected[i])
		}
	}
}

func TestUCS2EncodeNonMutation(t *testing.T) {
	input := []rune{0x61, 0x62, 0x63}
	original := [3]rune{0x61, 0x62, 0x63}
	got := UCS2Encode(input)
	if got != "abc" {
		t.Errorf("UCS2Encode([abc]) = %q, want %q", got, "abc")
	}
	for i, v := range original {
		if input[i] != v {
			t.Errorf("UCS2Encode mutated input[%d]: got %U, want %U", i, input[i], v)
		}
	}
}
