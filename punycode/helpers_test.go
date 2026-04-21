package punycode

import "testing"

// --- basicToDigit ---

func TestBasicToDigit(t *testing.T) {
	tests := []struct {
		cp   rune
		want int
	}{
		// Digits 0-9 → 26-35
		{'0', 26}, {'1', 27}, {'5', 31}, {'9', 35},
		// Uppercase A-Z → 0-25
		{'A', 0}, {'B', 1}, {'M', 12}, {'Z', 25},
		// Lowercase a-z → 0-25 (same values as uppercase)
		{'a', 0}, {'b', 1}, {'m', 12}, {'z', 25},
		// Boundaries that should return base (36)
		{0x2F, base}, // '/' — just below '0'
		{0x3A, base}, // ':' — just above '9'
		{0x40, base}, // '@' — just below 'A'
		{0x5B, base}, // '[' — just above 'Z'
		{0x60, base}, // '`' — just below 'a'
		{0x7B, base}, // '{' — just above 'z'
		{0x80, base}, // first non-ASCII
	}
	for _, tc := range tests {
		got := basicToDigit(tc.cp)
		if got != tc.want {
			t.Errorf("basicToDigit(%#x) = %d, want %d", tc.cp, got, tc.want)
		}
	}
}

// --- digitToBasic ---

func TestDigitToBasicLowercase(t *testing.T) {
	// flag=false: digits 0-25 → 'a'-'z', digits 26-35 → '0'-'9'
	for d := 0; d <= 25; d++ {
		want := rune('a' + d)
		if got := digitToBasic(d, false); got != want {
			t.Errorf("digitToBasic(%d, false) = %c, want %c", d, got, want)
		}
	}
	for d := 26; d <= 35; d++ {
		want := rune('0' + d - 26)
		if got := digitToBasic(d, false); got != want {
			t.Errorf("digitToBasic(%d, false) = %c, want %c", d, got, want)
		}
	}
}

func TestDigitToBasicUppercase(t *testing.T) {
	// flag=true: digits 0-25 → 'A'-'Z'
	for d := 0; d <= 25; d++ {
		want := rune('A' + d)
		if got := digitToBasic(d, true); got != want {
			t.Errorf("digitToBasic(%d, true) = %c, want %c", d, got, want)
		}
	}
}

func TestDigitToBasicRoundTrip(t *testing.T) {
	// digitToBasic then basicToDigit should be identity for 0..35
	for d := 0; d <= 35; d++ {
		cp := digitToBasic(d, false)
		if got := basicToDigit(cp); got != d {
			t.Errorf("roundtrip digit %d: digitToBasic → %c → basicToDigit = %d", d, cp, got)
		}
	}
}

// --- adapt ---

// Hand-verified vectors (computed by tracing the algorithm):
//
// adapt(900, 2, false):
//   delta = 900>>1=450; delta+=450/2=225 → 675
//   675>455: delta=floor(675/35)=19, k=36; 19<=455 exit
//   return 36 + 36*19/(19+38) = 36 + 684/57 = 36+12 = 48
//
// adapt(456, 1, false):
//   delta=456>>1=228; delta+=228/1=228 → 456
//   456>455: delta=floor(456/35)=13, k=36; 13<=455 exit
//   return 36 + 36*13/(13+38) = 36 + 468/51 = 36+9 = 45
//
// adapt(700, 1, true):
//   delta=floor(700/700)=1; delta+=1/1=1 → 2
//   2<=455, k=0
//   return 0 + 36*2/(2+38) = 72/40 = 1
//
// adapt(0, 1, false):
//   delta=0>>1=0; delta+=0 → 0; 0<=455; return 36*0/(0+38)=0
//
// adapt(455, 2, false):
//   delta=455>>1=227; delta+=227/2=113 → 340; 340<=455; k=0
//   return 36*340/(340+38)=12240/378=32
func TestAdapt(t *testing.T) {
	tests := []struct {
		delta, numPoints int
		firstTime        bool
		want             int
	}{
		{900, 2, false, 48},
		{456, 1, false, 45},
		{700, 1, true, 1},
		{0, 1, false, 0},
		{455, 2, false, 32},
	}
	for _, tc := range tests {
		got := adapt(tc.delta, tc.numPoints, tc.firstTime)
		if got != tc.want {
			t.Errorf("adapt(%d, %d, %v) = %d, want %d", tc.delta, tc.numPoints, tc.firstTime, got, tc.want)
		}
	}
}
