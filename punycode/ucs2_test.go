package punycode

import "testing"

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
	for _, v := range testVectors.UCS2 {
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
	for i, v := range testVectors.UCS2 {
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
