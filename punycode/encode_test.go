package punycode

import "testing"

func TestEncode(t *testing.T) {
	for _, v := range testVectors.Strings {
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
