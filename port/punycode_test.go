package punycode

import (
	"reflect"
	"testing"
)

// --- UCS2Decode ---

func TestUCS2Decode(t *testing.T) {
	for _, tc := range ucs2Data {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got := UCS2Decode(tc.encoded)
			if !reflect.DeepEqual(got, tc.decoded) {
				t.Errorf("UCS2Decode(%v)\n got  %v\n want %v", tc.encoded, got, tc.decoded)
			}
		})
	}
}

// --- UCS2Encode ---

func TestUCS2Encode(t *testing.T) {
	for _, tc := range ucs2Data {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got := UCS2Encode(tc.decoded)
			if !reflect.DeepEqual(got, tc.encoded) {
				t.Errorf("UCS2Encode(%v)\n got  %v\n want %v", tc.decoded, got, tc.encoded)
			}
		})
	}

	t.Run("does not mutate argument slice", func(t *testing.T) {
		in := []int32{0x61, 0x62, 0x63}
		orig := []int32{0x61, 0x62, 0x63}
		got := UCS2Encode(in)
		want := []uint16{0x61, 0x62, 0x63}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("UCS2Encode got %v, want %v", got, want)
		}
		if !reflect.DeepEqual(in, orig) {
			t.Errorf("input was mutated: got %v", in)
		}
	})
}

// --- Decode ---

func TestDecode(t *testing.T) {
	for _, tc := range stringsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.encoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := Decode(tc.encoded)
			if err != nil {
				t.Fatalf("Decode(%q) error: %v", tc.encoded, err)
			}
			if got != tc.decoded {
				t.Errorf("Decode(%q)\n got  %q\n want %q", tc.encoded, got, tc.decoded)
			}
		})
	}

	t.Run("handles uppercase Z (tests/tests.js:299)", func(t *testing.T) {
		got, err := Decode("ZZZ")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "\u7BA5" {
			t.Errorf("Decode(ZZZ) = %q, want %q", got, "\u7BA5")
		}
	})

	t.Run("ErrNotBasic on \\x81- (tests/tests.js:258)", func(t *testing.T) {
		_, err := Decode("\x81-")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("error on \\x81 (tests/tests.js:266)", func(t *testing.T) {
		_, err := Decode("\x81")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("ErrInvalidInput on ls8h= (tests/tests.js:306)", func(t *testing.T) {
		_, err := Decode("ls8h=")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

// --- Encode ---

func TestEncode(t *testing.T) {
	for _, tc := range stringsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.decoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := Encode(tc.decoded)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", tc.decoded, err)
			}
			if got != tc.encoded {
				t.Errorf("Encode(%q)\n got  %q\n want %q", tc.decoded, got, tc.encoded)
			}
		})
	}
}

// --- ToUnicode ---

func TestToUnicode(t *testing.T) {
	for _, tc := range domainsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.encoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := ToUnicode(tc.encoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", tc.encoded, err)
			}
			if got != tc.decoded {
				t.Errorf("ToUnicode(%q)\n got  %q\n want %q", tc.encoded, got, tc.decoded)
			}
		})
	}

	// Identity: strings not starting with xn-- pass through unchanged (tests/tests.js:333-343)
	for _, tc := range stringsData {
		tc := tc
		t.Run("identity/encoded/"+tc.encoded, func(t *testing.T) {
			got, err := ToUnicode(tc.encoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", tc.encoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToUnicode(%q) = %q, want unchanged", tc.encoded, got)
			}
		})
		t.Run("identity/decoded/"+tc.decoded, func(t *testing.T) {
			got, err := ToUnicode(tc.decoded)
			if err != nil {
				t.Fatalf("ToUnicode(%q) error: %v", tc.decoded, err)
			}
			if got != tc.decoded {
				t.Errorf("ToUnicode(%q) = %q, want unchanged", tc.decoded, got)
			}
		})
	}
}

// --- ToASCII ---

func TestToASCII(t *testing.T) {
	for _, tc := range domainsData {
		tc := tc
		name := tc.description
		if name == "" {
			name = tc.decoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := ToASCII(tc.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", tc.decoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToASCII(%q)\n got  %q\n want %q", tc.decoded, got, tc.encoded)
			}
		})
	}

	// Identity: already-ASCII strings survive round-trip (tests/tests.js:356-362)
	for _, tc := range stringsData {
		tc := tc
		t.Run("ascii-passthrough/"+tc.encoded, func(t *testing.T) {
			got, err := ToASCII(tc.encoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", tc.encoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToASCII(%q) = %q, want unchanged", tc.encoded, got)
			}
		})
	}

	// IDNA2003 separator normalisation (tests/tests.js:364-370)
	for _, tc := range separatorsData {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			got, err := ToASCII(tc.decoded)
			if err != nil {
				t.Fatalf("ToASCII(%q) error: %v", tc.decoded, err)
			}
			if got != tc.encoded {
				t.Errorf("ToASCII(%q)\n got  %q\n want %q", tc.decoded, got, tc.encoded)
			}
		})
	}
}

// --- Bootstring primitives unit tests ---

func TestBasicToDigit(t *testing.T) {
	cases := []struct{ in, want int32 }{
		{'a', 0}, {'z', 25}, {'A', 0}, {'Z', 25},
		{'0', 26}, {'9', 35},
		{'!', base}, {0x80, base},
	}
	for _, tc := range cases {
		if got := basicToDigit(tc.in); got != tc.want {
			t.Errorf("basicToDigit(%c/%d) = %d, want %d", tc.in, tc.in, got, tc.want)
		}
	}
}

func TestDigitToBasic(t *testing.T) {
	if got := digitToBasic(0, false); got != 'a' {
		t.Errorf("digitToBasic(0,false) = %c, want 'a'", got)
	}
	if got := digitToBasic(25, false); got != 'z' {
		t.Errorf("digitToBasic(25,false) = %c, want 'z'", got)
	}
	if got := digitToBasic(26, false); got != '0' {
		t.Errorf("digitToBasic(26,false) = %c, want '0'", got)
	}
	if got := digitToBasic(35, false); got != '9' {
		t.Errorf("digitToBasic(35,false) = %c, want '9'", got)
	}
}

func TestAdapt(t *testing.T) {
	// From the ZZZ→\u7BA5 trace: adapt(31525,1,true)=25
	if got := adapt(31525, 1, true); got != 25 {
		t.Errorf("adapt(31525,1,true) = %d, want 25", got)
	}
}
