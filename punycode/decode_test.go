package punycode

import (
	"testing"
)

func TestDecode(t *testing.T) {
	for _, v := range testVectors.Strings {
		name := v.description
		if name == "" {
			name = v.encoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := Decode(v.encoded)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", v.encoded, err)
			}
			if got != v.decoded {
				t.Errorf("Decode(%q)\n got: %q\nwant: %q", v.encoded, got, v.decoded)
			}
		})
	}
}

func TestDecodeUppercase(t *testing.T) {
	// tests/tests.js:299-301: case-insensitive base-36 decoding.
	got, err := Decode("ZZZ")
	if err != nil {
		t.Fatalf("Decode(\"ZZZ\") error: %v", err)
	}
	want := "\u7BA5"
	if got != want {
		t.Errorf("Decode(\"ZZZ\") = %q, want %q", got, want)
	}
}

func TestDecodeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  error
	}{
		{
			// Non-basic code point in the basic prefix (tests/tests.js:255-261).
			name:  "not-basic prefix",
			input: "\x81-",
			want:  ErrNotBasic,
		},
		{
			// "\x81" has no delimiter so the main loop tries to decode it;
			// basicToDigit(0x81) == base, triggering ErrInvalidInput.
			// The JS test description says "Overflow" but the actual thrown
			// error is "Invalid input" — the description is misleading.
			name:  "non-basic sole char",
			input: "\x81",
			want:  ErrInvalidInput,
		},
		{
			// '=' (0x3D) is not a valid base-36 digit (tests/tests.js:302-309).
			name:  "invalid base-36 digit",
			input: "ls8h=",
			want:  ErrInvalidInput,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Decode(tt.input)
			if err != tt.want {
				t.Errorf("Decode(%q) = %v, want %v", tt.input, err, tt.want)
			}
		})
	}
}
