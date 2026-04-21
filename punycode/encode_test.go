package punycode

import "testing"

// encodeVectors are the same 24 string fixtures used by TestDecode,
// run in reverse: input=decoded, expected=encoded.
// Source: tests/tests.js:7-136 (testData.strings).
var encodeVectors = decodeVectors

func TestEncode(t *testing.T) {
	for _, v := range encodeVectors {
		name := v.description
		if name == "" {
			name = v.decoded
		}
		t.Run(name, func(t *testing.T) {
			got, err := Encode(v.decoded)
			if err != nil {
				t.Fatalf("Encode(%q) error: %v", v.decoded, err)
			}
			if got != v.encoded {
				t.Errorf("Encode(%q)\n got: %q\nwant: %q", v.decoded, got, v.encoded)
			}
		})
	}
}
