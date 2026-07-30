// Package diff cross-validates twister against snare: every edit-based fuzzer
// makes exactly one Damerau-Levenshtein edit, so snare must detect each generated
// variant as a near-miss of the seed at distance 1. This proves both libraries at
// once — twister generates the candidate, snare confirms it is one edit away.
package diff

import (
	"testing"

	"github.com/netstar-labs/snare"
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
		set := snare.New([]string{seed})
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

// aggressive2 returns the aggressive (2-substitution) variants of seed, asserting
// each is tagged EditCount 2 — the multi-edit class the round-trip below is about.
func aggressive2(t *testing.T, seed string) []string {
	t.Helper()
	var out []string
	for _, v := range twister.PermuteWith(seed, twister.Options{Fuzzers: []string{"aggressive"}, MaxEdits: 2}) {
		if v.EditCount != 2 {
			t.Errorf("aggressive %q of %q has EditCount %d, want 2", v.Name, seed, v.EditCount)
		}
		out = append(out, v.Name)
	}
	if len(out) == 0 {
		t.Fatalf("no aggressive variants generated for %q", seed)
	}
	return out
}

// TestMultiEditNotDetectedAtDistance1 is the per-class complement of
// TestGenerateThenDetect: multi-edit variants are deliberately excluded from the
// "every variant is snare@1" contract, so this documents where they actually land.
//
// snare's edit cap is length-relative (k=1 for queries of <= 6 runes, k=2 above), so
// an aggressive 2-substitution variant is:
//   - a short brand (paypal, 6 runes): outside k=1 entirely — not detected; and
//   - a long brand (microsoft, 9 runes): detected, but at distance 2, never 1.
//
// In neither case does it round-trip at snare@1 — which is exactly why the core diff
// harness must not assert it does.
func TestMultiEditNotDetectedAtDistance1(t *testing.T) {
	// Short seed: 2-edit variants fall outside snare's k=1 budget → not a hit.
	const shortSeed = "paypal"
	shortSet := snare.New([]string{shortSeed})
	for _, name := range aggressive2(t, shortSeed) {
		if got, dist, ok := shortSet.Nearest(name); ok && dist == 1 {
			t.Errorf("aggressive 2-sub %q of %q was detected at snare@1 (%q, dist %d); the multi-edit class must not round-trip at distance 1",
				name, shortSeed, got, dist)
		}
	}

	// Long seed: k=2, so the class IS recoverable — but at distance 2, not 1.
	const longSeed = "microsoft"
	longSet := snare.New([]string{longSeed})
	sawDist2 := false
	for _, name := range aggressive2(t, longSeed) {
		got, dist, ok := longSet.Nearest(name)
		if ok && dist == 1 {
			t.Errorf("aggressive 2-sub %q of %q detected at snare@1; must be distance 2, not 1", name, longSeed)
		}
		if ok && got == longSeed && dist == 2 {
			sawDist2 = true
		}
	}
	if !sawDist2 {
		t.Errorf("expected at least one aggressive 2-sub variant of %q to be detected by snare at distance 2", longSeed)
	}
}
