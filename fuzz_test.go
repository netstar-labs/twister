package twister

import (
	"reflect"
	"strings"
	"testing"
)

// FuzzPermute asserts the invariants hold for arbitrary input: no panic,
// deterministic output, deduplicated, seed excluded, and every variant tagged.
func FuzzPermute(f *testing.F) {
	for _, s := range []string{"", "a", "go", "google", "paypal", "xn--80ak6aa92e", "über", strings.Repeat("a", 256)} {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, label string) {
		a := PermuteWith(label, Options{TLDs: []string{"com", "net"}})
		if b := PermuteWith(label, Options{TLDs: []string{"com", "net"}}); !reflect.DeepEqual(a, b) {
			t.Fatal("Permute is not deterministic")
		}

		norm := strings.ToLower(strings.TrimSpace(label))
		seen := make(map[string]bool, len(a))
		for i, v := range a {
			if v.Name == "" {
				t.Errorf("empty variant for %q", label)
			}
			if v.Name == norm {
				t.Errorf("seed %q not excluded", norm)
			}
			if seen[v.Name] {
				t.Errorf("duplicate variant %q", v.Name)
			}
			seen[v.Name] = true
			if v.Fuzzer == "" {
				t.Errorf("untagged variant %q", v.Name)
			}
			if i > 0 && a[i-1].Name > v.Name {
				t.Errorf("output not sorted at %d: %q > %q", i, a[i-1].Name, v.Name)
			}
		}
	})
}
