// Package diff cross-validates twister against twist: every edit-based fuzzer
// makes exactly one Damerau-Levenshtein edit, so twist must detect each generated
// variant as a near-miss of the seed at distance 1. This proves both libraries at
// once — twister generates the candidate, twist confirms it is one edit away.
package diff

import (
	"testing"

	"github.com/netstar-labs/twist"
	"github.com/netstar-labs/twister"
)

// editFuzzers are the techniques that make exactly one rune edit (insert, delete,
// substitute, or adjacent transpose) — so each variant is distance 1 from the
// seed. tld-swap (appends .tld) and the multi-emit data fuzzers are excluded only
// where they are not single-edit; homoglyph, leet, and bitsquatting are single rune
// substitutions and belong here too.
var editFuzzers = []string{
	"omission", "repetition", "transposition", "replacement", "insertion",
	"addition", "hyphenation", "subdomain", "vowel-swap", "homoglyph", "leet", "bitsquatting",
}

func TestGenerateThenDetect(t *testing.T) {
	seeds := []string{"paypal", "google", "amazon", "microsoft", "cloudflare"}
	for _, seed := range seeds {
		set := twist.New([]string{seed})
		for _, f := range editFuzzers {
			for _, v := range twister.PermuteWith(seed, twister.Options{Fuzzers: []string{f}}) {
				got, dist, ok := set.Nearest(v.Name)
				if !ok || got != seed || dist != 1 {
					t.Errorf("%s/%q: Nearest(%q) = (%q, %d, %v), want (%q, 1, true)",
						f, seed, v.Name, got, dist, ok, seed)
				}
			}
		}
	}
}
