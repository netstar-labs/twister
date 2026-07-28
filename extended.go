package twister

import (
	"sort"
	"strings"
	"unicode/utf8"
)

// The fuzzers below are the v0.2 multi-edit extensions. Unlike the single-edit
// core in fuzzers.go, each makes more than one Damerau-Levenshtein edit, so their
// output never belongs in Permute's dist-1 set. They are OFF by default and behave
// like tld-swap: combosquat and aggressive sit in [FuzzerNames] but stay silent
// until their Options field is set, while multi-homoglyph and homophone emit
// unconditionally and so are kept out of [FuzzerNames] (see [ExtendedFuzzerNames])
// to keep Permute dist-1 — a caller runs them by naming them in [Options.Fuzzers].
//
// A generator here returns rawVariant values rather than bare strings so each
// carries its own edit count; the dispatcher ([PermuteWith]) dedups them, drops the
// seed, and tags each with the fuzzer name.

// rawVariant is a generated label plus the number of edits it lies from the seed —
// the multi-edit fuzzers' output before dedup and tagging.
type rawVariant struct {
	name  string
	edits int
}

// extRegistry maps each multi-edit technique to its generator. It is consulted
// after [registry]: a name resolves to a single-edit core fuzzer first, then to an
// extension here.
var extRegistry = map[string]func(r []rune, o Options) []rawVariant{
	"combosquat":      fuzzCombosquat,
	"aggressive":      fuzzAggressive,
	"multi-homoglyph": fuzzMultiHomoglyph,
	"homophone":       fuzzHomophone,
}

// MaxAggressiveVariants is the exported, deterministic upper bound on how many
// variants [fuzzAggressive] emits for one seed: its candidate space is
// (positions choose k) × (candidates per position)^k and blows up with MaxEdits, so
// generation stops here. It is a public constant rather than a side-effecting log line
// — a library must not write to the global logger — so a caller that reaches the cap
// gets a truncated-but-deterministic set and can compare len(output) against this
// bound. Sized to admit a full 2-edit sweep of a normal brand while capping the
// higher-k explosions.
const MaxAggressiveVariants = 10000

// fuzzCombosquat affixes each caller-supplied keyword into and around the brand:
// brand-word, word-brand, brandword, wordbrand, and brand.word. It is the dominant
// real-world phishing class and a composition, not an edit — detected by token
// containment / n-gram similarity, not twist — so every variant is tagged with the
// count of runes it adds (>1 for any real keyword). Off unless [Options.Words] is
// set, exactly like tld-swap is off unless [Options.TLDs] is set.
func fuzzCombosquat(r []rune, o Options) []rawVariant {
	if len(o.Words) == 0 {
		return nil
	}
	brand := string(r)
	n := len(r)
	out := make([]rawVariant, 0, len(o.Words)*5)
	for _, w := range o.Words {
		w = strings.ToLower(strings.TrimSpace(w))
		if w == "" {
			continue
		}
		for _, f := range []string{
			brand + "-" + w, // brand-word
			w + "-" + brand, // word-brand
			brand + w,       // brandword
			w + brand,       // wordbrand
			brand + "." + w, // brand.word
		} {
			out = append(out, rawVariant{name: f, edits: utf8.RuneCountInString(f) - n})
		}
	}
	return out
}

// fuzzAggressive composes the substitution fuzzers (leet, homoglyph, keyboard
// replacement) at up to [Options.MaxEdits] positions at once — p4yp4l, g00gl3, and
// their multi-homoglyph cousins. It enumerates every k-combination of positions for
// k in 2..MaxEdits and every product of per-position candidates, deduped by the
// dispatcher, bounded by [MaxAggressiveVariants]. Off when MaxEdits < 2.
//
// Each variant is tagged with its true Damerau-Levenshtein (OSA) distance from the
// seed — not the substitution count — because two adjacent substitutions can
// coincide with a single transposition (subbing f→t and t→f is one swap, distance
// 1, not two edits). Such collapses belong to the core transposition fuzzer, so they
// are dropped here; every surviving aggressive variant is a genuine 2+-edit look-alike.
func fuzzAggressive(r []rune, o Options) []rawVariant {
	if o.MaxEdits < 2 {
		return nil
	}
	// Per-position substitution candidates, deterministic order, original rune
	// excluded; positions with no candidate are skipped so combinations stay dense.
	var positions []int
	cands := make(map[int][]rune, len(r))
	for i := range r {
		if c := substitutionCandidates(r[i]); len(c) > 0 {
			positions = append(positions, i)
			cands[i] = c
		}
	}
	maxK := o.MaxEdits
	if maxK > len(positions) {
		maxK = len(positions)
	}

	var out []rawVariant
	capped := false
	for k := 2; k <= maxK && !capped; k++ {
		forEachCombination(len(positions), k, func(combo []int) bool {
			pos := make([]int, k)
			lists := make([][]rune, k)
			for j, ci := range combo {
				pos[j] = positions[ci]
				lists[j] = cands[positions[ci]]
			}
			return forEachProduct(lists, func(choice []rune) bool {
				vr := make([]rune, len(r))
				copy(vr, r)
				for j, p := range pos {
					vr[p] = choice[j]
				}
				if d := osaDistance(r, vr); d >= 2 {
					out = append(out, rawVariant{name: string(vr), edits: d})
					if len(out) >= MaxAggressiveVariants {
						capped = true
						return false
					}
				}
				return true
			})
		})
	}
	return out
}

// fuzzMultiHomoglyph applies each multi-rune confusable rule (see [multiHomoglyphs])
// at every matching position: rn→m, vv→w, cl→d, nn→m, and the m→rn reverse. Each is
// two edits (a substitution plus an insertion or deletion), so it is excluded from
// the single-rune homoglyph core. Detected by skeleton equality, not twist@1.
func fuzzMultiHomoglyph(r []rune, _ Options) []rawVariant {
	var out []rawVariant
	for _, key := range sortedStringKeys(multiHomoglyphs) {
		kr := []rune(key)
		for i := 0; i+len(kr) <= len(r); i++ {
			if !runesEqualAt(r, i, kr) {
				continue
			}
			for _, rep := range multiHomoglyphs[key] {
				out = append(out, rawVariant{
					name:  replaceRange(r, i, len(kr), []rune(rep)),
					edits: editDistance(key, rep),
				})
			}
		}
	}
	return out
}

// fuzzHomophone substitutes sound-alike substrings (ph↔f, c↔k, s↔z, ck↔k) at every
// matching position, one rewrite per variant. The embedded [homophones] table is the
// default; [Options.Homophones] overrides it per key. Each variant is tagged with the
// true edit cost of its rewrite (ph↔f is 2, c↔k is 1). The class is defined by
// phonetic equality, so it routes to a phonetic detector rather than twist@1.
func fuzzHomophone(r []rune, o Options) []rawVariant {
	table := homophones
	if len(o.Homophones) > 0 {
		table = mergeHomophones(homophones, o.Homophones)
	}
	var out []rawVariant
	for _, key := range sortedStringKeys(table) {
		kr := []rune(key)
		for i := 0; i+len(kr) <= len(r); i++ {
			if !runesEqualAt(r, i, kr) {
				continue
			}
			for _, rep := range table[key] {
				out = append(out, rawVariant{
					name:  replaceRange(r, i, len(kr), []rune(rep)),
					edits: editDistance(key, rep),
				})
			}
		}
	}
	return out
}

// --- extended-fuzzer helpers ----------------------------------------------------

// substitutionCandidates returns the deduped substitution look-alikes for c — the
// union of its leet numeral, single-rune homoglyphs, and keyboard neighbours, in
// that fixed order — with c itself excluded. The order is deterministic so the
// aggressive fuzzer's enumeration (and its cap) is reproducible.
func substitutionCandidates(c rune) []rune {
	var out []rune
	seen := map[rune]bool{c: true}
	add := func(cs []rune) {
		for _, x := range cs {
			if !seen[x] {
				seen[x] = true
				out = append(out, x)
			}
		}
	}
	add(leet[c])
	add(homoglyphs[c])
	add(keyboard[c])
	return out
}

// forEachCombination calls fn with each k-combination of the indices [0,n) in
// lexicographic order. fn returns false to stop early, and forEachCombination then
// returns false so an outer loop can stop too. combo is reused between calls; fn
// must not retain it.
func forEachCombination(n, k int, fn func(combo []int) bool) bool {
	if k <= 0 || k > n {
		return true
	}
	combo := make([]int, k)
	for i := range combo {
		combo[i] = i
	}
	for {
		if !fn(combo) {
			return false
		}
		i := k - 1
		for i >= 0 && combo[i] == n-k+i {
			i--
		}
		if i < 0 {
			return true
		}
		combo[i]++
		for j := i + 1; j < k; j++ {
			combo[j] = combo[j-1] + 1
		}
	}
}

// forEachProduct calls fn with each element of the Cartesian product of lists (an
// odometer over one choice per list), in a deterministic order. fn returns false to
// stop early. choice is reused between calls; fn must not retain it. Every list is
// assumed non-empty.
func forEachProduct(lists [][]rune, fn func(choice []rune) bool) bool {
	idx := make([]int, len(lists))
	choice := make([]rune, len(lists))
	for {
		for i := range lists {
			choice[i] = lists[i][idx[i]]
		}
		if !fn(choice) {
			return false
		}
		p := len(lists) - 1
		for p >= 0 {
			idx[p]++
			if idx[p] < len(lists[p]) {
				break
			}
			idx[p] = 0
			p--
		}
		if p < 0 {
			return true
		}
	}
}

// replaceRange returns r with the n runes at start replaced by repl, leaving r
// untouched.
func replaceRange(r []rune, start, n int, repl []rune) string {
	out := make([]rune, 0, len(r)-n+len(repl))
	out = append(out, r[:start]...)
	out = append(out, repl...)
	out = append(out, r[start+n:]...)
	return string(out)
}

// runesEqualAt reports whether the runes of r starting at i equal seq.
func runesEqualAt(r []rune, i int, seq []rune) bool {
	for j, c := range seq {
		if r[i+j] != c {
			return false
		}
	}
	return true
}

// sortedStringKeys returns the keys of m in ascending order, so a fuzzer that ranges
// a table produces deterministic output regardless of Go's map iteration order.
func sortedStringKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// mergeHomophones overlays extra onto base, extra winning per key, without mutating
// either — the merge behind [Options.Homophones] extending the embedded defaults.
func mergeHomophones(base, extra map[string][]string) map[string][]string {
	out := make(map[string][]string, len(base)+len(extra))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// editDistance is the Levenshtein distance between two short strings — the honest
// edit cost of a homophone or multi-homoglyph rewrite rule, used to tag its variant.
// The rules are a few runes each, so the full row-swapped DP is trivial.
func editDistance(a, b string) int {
	ar, br := []rune(a), []rune(b)
	prev := make([]int, len(br)+1)
	curr := make([]int, len(br)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ar); i++ {
		curr[0] = i
		for j := 1; j <= len(br); j++ {
			cost := 1
			if ar[i-1] == br[j-1] {
				cost = 0
			}
			curr[j] = min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(br)]
}

// osaDistance is the optimal-string-alignment (restricted Damerau-Levenshtein)
// distance over runes — the same metric the sibling twist uses. It counts an
// adjacent transposition as one edit, so the aggressive fuzzer can tell a genuine
// two-substitution variant (distance 2) from two substitutions that happen to swap
// an adjacent pair (distance 1). Three-row DP; used only by the opt-in aggressive
// fuzzer, so its O(len(a)*len(b)) cost is off the core path.
func osaDistance(a, b []rune) int {
	lb := len(b)
	prev2 := make([]int, lb+1) // row i-2
	prev := make([]int, lb+1)  // row i-1
	curr := make([]int, lb+1)  // row i
	for j := 0; j <= lb; j++ {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			v := min(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
			if i > 1 && j > 1 && a[i-1] == b[j-2] && a[i-2] == b[j-1] {
				v = min(v, prev2[j-2]+1)
			}
			curr[j] = v
		}
		prev2, prev, curr = prev, curr, prev2
	}
	return prev[lb]
}
