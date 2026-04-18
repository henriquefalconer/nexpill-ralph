package punycode

import "testing"

// TestConstants asserts each Bootstring constant matches RFC 3492 §5
// and the corresponding value in punycode.js:3-14.
func TestConstants(t *testing.T) {
	cases := []struct {
		name string
		got  int32
		want int32
	}{
		{"maxInt", maxInt, 2147483647},
		{"base", base, 36},
		{"tMin", tMin, 1},
		{"tMax", tMax, 26},
		{"skew", skew, 38},
		{"damp", damp, 700},
		{"initialBias", initialBias, 72},
		{"initialN", initialN, 128},
		{"baseMinusTMin", baseMinusTMin, 35},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s = %d, want %d", c.name, c.got, c.want)
		}
	}
	if delimiter != '-' {
		t.Errorf("delimiter = %q, want '-'", delimiter)
	}
}

// TestErrors asserts sentinel error messages match punycode.js:23-25 verbatim.
func TestErrors(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"ErrOverflow", ErrOverflow, "Overflow: input needs wider integers to process"},
		{"ErrNotBasic", ErrNotBasic, "Illegal input >= 0x80 (not a basic code point)"},
		{"ErrInvalidInput", ErrInvalidInput, "Invalid input"},
	}
	for _, c := range cases {
		if c.err.Error() != c.want {
			t.Errorf("%s message = %q, want %q", c.name, c.err.Error(), c.want)
		}
	}
}
