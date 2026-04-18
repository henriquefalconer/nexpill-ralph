package punycode

import "testing"

func TestMapDomain(t *testing.T) {
	identity := func(s string) (string, error) { return s, nil }

	t.Run("plain domain", func(t *testing.T) {
		got, err := mapDomain("example.com", identity)
		if err != nil || got != "example.com" {
			t.Fatalf("got %q, err %v", got, err)
		}
	})

	t.Run("email splitting — only domain part passed to callback", func(t *testing.T) {
		var seen string
		_, err := mapDomain("foo@bar.com", func(label string) (string, error) {
			seen += label + ";"
			return label, nil
		})
		if err != nil {
			t.Fatal(err)
		}
		if seen != "bar;com;" {
			t.Fatalf("labels seen: %q, want %q", seen, "bar;com;")
		}
	})

	t.Run("email prefix preserved in output", func(t *testing.T) {
		got, err := mapDomain("foo@bar.com", identity)
		if err != nil || got != "foo@bar.com" {
			t.Fatalf("got %q, err %v", got, err)
		}
	})

	// Separator normalisation: U+3002, U+FF0E, U+FF61 → U+002E. tests/tests.js:221-242
	for _, sep := range separatorFixtures[1:] {
		sep := sep
		t.Run(sep.description, func(t *testing.T) {
			// After normalisation the separators become "." so split yields two labels.
			got, err := mapDomain(sep.decoded, identity)
			if err != nil {
				t.Fatal(err)
			}
			want := "ma\u00F1ana.com"
			if got != want {
				t.Fatalf("got %q, want %q", got, want)
			}
		})
	}
}
