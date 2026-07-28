package twister

import (
	"sort"
	"strings"
)

// Variant is a generated permutation of a seed label, tagged with the fuzzer
// technique that produced it.
type Variant struct {
	Name   string // the permuted label (a bare label, or label.tld for tld-swap)
	Fuzzer string // the technique: "omission", "transposition", "homoglyph", …
}

// Options selects which fuzzers run and supplies caller-owned data.
type Options struct {
	// Fuzzers restricts generation to the named techniques (see [FuzzerNames]).
	// Empty runs every fuzzer. Unknown names are ignored.
	Fuzzers []string
	// TLDs is the list the tld-swap fuzzer swaps in (e.g. {"com","net","org"});
	// twister ships no TLD list of its own. Every other fuzzer ignores it.
	TLDs []string
}

// FuzzerNames lists the technique names in the deterministic order they run. When
// two techniques would produce the same label, the earlier one in this order owns
// the deduplicated variant. The keyboard-adjacency table (docs call it the
// "keyboard" fuzzer) is not a standalone entry — it powers insertion and
// replacement.
var FuzzerNames = []string{
	"omission",
	"repetition",
	"transposition",
	"replacement",
	"insertion",
	"addition",
	"hyphenation",
	"subdomain",
	"vowel-swap",
	"homoglyph",
	"leet",
	"bitsquatting",
	"tld-swap",
}

// registry maps each technique to its generator. A generator returns raw variant
// labels (possibly with duplicates, or the seed); the dispatcher dedups and drops
// the seed.
var registry = map[string]func(r []rune, o Options) []string{
	"omission":      fuzzOmission,
	"repetition":    fuzzRepetition,
	"transposition": fuzzTransposition,
	"replacement":   fuzzReplacement,
	"insertion":     fuzzInsertion,
	"addition":      fuzzAddition,
	"hyphenation":   fuzzHyphenation,
	"subdomain":     fuzzSubdomain,
	"vowel-swap":    fuzzVowelSwap,
	"homoglyph":     fuzzHomoglyph,
	"leet":          fuzzLeet,
	"bitsquatting":  fuzzBitsquatting,
	"tld-swap":      fuzzTLDSwap,
}

// Permute generates the full core permutation set for label, deduplicated and
// excluding the seed. The tld-swap fuzzer yields nothing here (it has no TLD
// list); use [PermuteWith] with [Options.TLDs] to enable it.
func Permute(label string) []Variant {
	return PermuteWith(label, Options{})
}

// PermuteWith generates permutations under o: [Options.Fuzzers] selects the
// techniques (empty = all), [Options.TLDs] feeds tld-swap. The label is
// lower-cased and trimmed first. The result is deduplicated, excludes the seed,
// and is sorted by name for determinism.
func PermuteWith(label string, o Options) []Variant {
	label = strings.ToLower(strings.TrimSpace(label))
	if label == "" {
		return nil
	}
	names := o.Fuzzers
	if len(names) == 0 {
		names = FuzzerNames
	}
	r := []rune(label)

	// Rough pre-size to the expected variant count (a fixed base from the
	// length-independent fuzzers plus ~26/rune), avoiding repeated map rehash and
	// slice grow. Measured fit: go→68, google→188, verylongbrandname→470.
	est := 32 + len(r)*26
	seen := make(map[string]struct{}, est+1)
	seen[label] = struct{}{} // exclude the seed
	out := make([]Variant, 0, est)
	for _, name := range names {
		fn, ok := registry[name]
		if !ok {
			continue
		}
		for _, v := range fn(r, o) {
			if !validLabel(v) {
				continue
			}
			if _, dup := seen[v]; dup {
				continue
			}
			seen[v] = struct{}{}
			out = append(out, Variant{Name: v, Fuzzer: name})
		}
	}
	if len(out) == 0 {
		return nil // no fuzzer produced anything (e.g. only unknown names)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// validLabel rejects the empty string and any label whose first or last character
// is a hyphen or dot — an unregistrable host label. Some fuzzers can otherwise
// emit one: bitsquatting can flip an ASCII byte to '-', and omission/transposition
// can expose a hyphen at a boundary of a seed that already contains one.
func validLabel(v string) bool {
	if v == "" {
		return false
	}
	first, last := v[0], v[len(v)-1]
	return first != '-' && first != '.' && last != '-' && last != '.'
}
