package punycode

import "testing"

// TestBasicToDigit covers every branch boundary from specs/src-punycode.md:114-123.
// Also verifies case-insensitive folding required by tests/tests.js:299-301 (ZZZ→U+7BA5).
func TestBasicToDigit(t *testing.T) {
	tests := []struct {
		in   int32
		want int32
	}{
		// digits '0'..'9' → 26..35
		{'0', 26},
		{'9', 35},
		// uppercase 'A'..'Z' → 0..25
		{'A', 0},
		{'Z', 25},
		// lowercase 'a'..'z' → 0..25 (case-insensitive folding)
		{'a', 0},
		{'z', 25},
		// boundaries just outside valid ranges → base (36)
		{'/', 36},  // 0x2F, one below '0'
		{':', 36},  // 0x3A, one above '9'
		{'@', 36},  // 0x40, one below 'A'
		{'[', 36},  // 0x5B, one above 'Z'
		{'`', 36},  // 0x60, one below 'a'
		{'{', 36},  // 0x7B, one above 'z'
		{0x80, 36}, // non-basic
	}
	for _, tt := range tests {
		got := basicToDigit(tt.in)
		if got != tt.want {
			t.Errorf("basicToDigit(%#x) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

// TestDigitToBasic covers every digit 0..35.
// specs/src-punycode.md:126-138; punycode.js:157-172.
func TestDigitToBasic(t *testing.T) {
	for digit := int32(0); digit <= 35; digit++ {
		got := digitToBasic(digit)
		var want byte
		if digit < 26 {
			want = byte('a' + digit)
		} else {
			want = byte('0' + digit - 26)
		}
		if got != want {
			t.Errorf("digitToBasic(%d) = %q, want %q", digit, got, want)
		}
	}
}

// TestAdapt verifies the bias adaptation function.
// RFC 3492 §3.4; punycode.js:174-187; specs/src-punycode.md:142-150.
func TestAdapt(t *testing.T) {
	tests := []struct {
		delta, numPoints int32
		firstTime        bool
		want             int32
	}{
		// adapt(0, 1, true): delta/damp=0; 0+0/1=0; loop skips; 0+(36)*0/(0+38)=0
		{0, 1, true, 0},
		// adapt(700, 1, true): delta=700/700=1; 1+1/1=2; 2<=455 skip; 0+36*2/(2+38)=72/40=1
		{700, 1, true, 1},
		// adapt(700, 1, false): delta=350; 350+350=700; 700>455: delta=20,k=36; 20<=455; 36+36*20/(20+38)=36+720/58=36+12=48
		{700, 1, false, 48},
	}
	for _, tt := range tests {
		got := adapt(tt.delta, tt.numPoints, tt.firstTime)
		if got != tt.want {
			t.Errorf("adapt(%d, %d, %v) = %d, want %d", tt.delta, tt.numPoints, tt.firstTime, got, tt.want)
		}
	}
}
