package twister

import (
	"sort"
	"strings"
)

// Variant is a generated permutation of a seed label, tagged with the fuzzer
// technique that produced it and the number of edits it lies from the seed.
type Variant struct {
	Name   string // the permuted label (a bare label, or label.tld for tld-swap)
	Fuzzer string // the technique: "omission", "transposition", "homoglyph", …
	// EditCount is the Damerau-Levenshtein distance from the seed. Every single-edit
	// core fuzzer sets 1; the multi-edit extensions (tld-swap, combosquat, aggressive,
	// multi-homoglyph, homophone) set the true edit cost of the rewrite they applied —
	// usually >1, though a homophone swap like c→k is a single edit — so a consumer can
	// route each variant to the detector that can recover it.
	EditCount int
}

// Options selects which fuzzers run and supplies caller-owned data.
type Options struct {
	// Fuzzers restricts generation to the named techniques (see [FuzzerNames] and
	// [ExtendedFuzzerNames]). Empty runs every fuzzer in [FuzzerNames]; the
	// unconditional multi-edit extensions ([ExtendedFuzzerNames]) run only when named
	// here. Unknown names are ignored.
	Fuzzers []string
	// TLDs is the list the tld-swap fuzzer swaps in (e.g. {"com","net","org"});
	// twister ships no TLD list of its own. Every other fuzzer ignores it.
	TLDs []string
	// Words are the combosquat keywords affixed to the brand (e.g. {"login","secure"}
	// → paypal-login, secure-paypal, …); twister ships none. The combosquat fuzzer is
	// off unless this is set, mirroring tld-swap. Every other fuzzer ignores it.
	Words []string
	// MaxEdits caps how many positions the aggressive fuzzer substitutes at once
	// (p4yp4l is 2). 0 (the default) keeps aggressive off; a value < 2 yields nothing.
	// Every other fuzzer ignores it.
	MaxEdits int
	// Homophones overrides the embedded homophone table per key (see [homophones]);
	// nil uses the embedded defaults. Only the homophone fuzzer reads it.
	Homophones map[string][]string
}

// FuzzerNames lists the default-run technique names in the deterministic order they
// run. When two techniques would produce the same label, the earlier one in this
// order owns the deduplicated variant. The keyboard-adjacency table (docs call it
// the "keyboard" fuzzer) is not a standalone entry — it powers insertion and
// replacement.
//
// The first twelve are the single-edit core: each makes exactly one edit, so
// [Permute] over them is the tight dist-1 set. tld-swap, combosquat, and aggressive
// are multi-edit but data-gated — silent until [Options.TLDs] / [Options.Words] /
// [Options.MaxEdits] is set — so they are safe in the default list without breaking
// Permute's dist-1 invariant, and each carries its true edit count (they live in
// [extRegistry], not the EditCount-1 [registry]). The unconditional multi-edit
// extensions are in [ExtendedFuzzerNames] instead.
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
	"combosquat",
	"aggressive",
}

// ExtendedFuzzerNames lists the multi-edit techniques that emit variants
// unconditionally (from embedded tables, with no data gate). They are kept OUT of
// [FuzzerNames] so [Permute] — and any empty-[Options.Fuzzers] call — stays the
// tight dist-1 set; a caller runs them by naming them in [Options.Fuzzers].
var ExtendedFuzzerNames = []string{
	"multi-homoglyph",
	"homophone",
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
}

// Permute generates the single-edit core permutation set for label, deduplicated
// and excluding the seed — every variant is exactly one edit from the seed
// (EditCount 1). The data-gated fuzzers yield nothing here (no TLD, keyword, or
// MaxEdits supplied), and the unconditional multi-edit extensions
// ([ExtendedFuzzerNames]) do not run at all; use [PermuteWith] to enable them.
func Permute(label string) []Variant {
	return PermuteWith(label, Options{})
}

// PermuteWith generates permutations under o: [Options.Fuzzers] selects the
// techniques (empty runs [FuzzerNames]; the [ExtendedFuzzerNames] techniques run
// only when named), [Options.TLDs] feeds tld-swap, [Options.Words] feeds combosquat,
// [Options.MaxEdits] feeds aggressive, and [Options.Homophones] overrides the
// homophone table. The label is lower-cased and trimmed first. The result is
// deduplicated, excludes the seed, and is sorted by name for determinism.
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
	add := func(name, fuzzer string, edits int) {
		if !validLabel(name) {
			return
		}
		if _, dup := seen[name]; dup {
			return
		}
		seen[name] = struct{}{}
		out = append(out, Variant{Name: name, Fuzzer: fuzzer, EditCount: edits})
	}
	for _, name := range names {
		// A name resolves to a single-edit core fuzzer first (EditCount 1), then to a
		// multi-edit extension (which carries its own per-variant edit count).
		if fn, ok := registry[name]; ok {
			for _, v := range fn(r, o) {
				add(v, name, 1)
			}
			continue
		}
		if fn, ok := extRegistry[name]; ok {
			for _, rv := range fn(r, o) {
				add(rv.name, name, rv.edits)
			}
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
