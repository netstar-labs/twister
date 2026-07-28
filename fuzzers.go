package twister

import "strings"

// The fuzzers below each take the seed's runes and return raw variant labels; the
// dispatcher ([PermuteWith]) dedups them and drops the seed. Every edit-based
// fuzzer makes exactly one Damerau-Levenshtein edit, so each variant is one step
// from the seed. Output size is linear in label length per fuzzer.

// fuzzOmission drops each character in turn (google → oogle, gogle, goole, …).
func fuzzOmission(r []rune, _ Options) []string {
	if len(r) < 2 {
		return nil // dropping the only rune yields the empty label
	}
	out := make([]string, 0, len(r))
	for i := range r {
		out = append(out, withDeleted(r, i))
	}
	return out
}

// fuzzRepetition doubles each character (google → ggoogle, gooogle, …).
func fuzzRepetition(r []rune, _ Options) []string {
	out := make([]string, 0, len(r))
	for i := range r {
		out = append(out, withInserted(r, i, r[i]))
	}
	return out
}

// fuzzTransposition swaps each adjacent pair (google → ogogle, gogole, …). Equal
// neighbours are skipped — swapping them reproduces the seed.
func fuzzTransposition(r []rune, _ Options) []string {
	out := make([]string, 0, len(r))
	for i := 0; i+1 < len(r); i++ {
		if r[i] == r[i+1] {
			continue
		}
		out = append(out, withSwapped(r, i, i+1))
	}
	return out
}

// eachReplacement emits withReplaced at every position for each candidate rune
// that cand returns for the rune there — the shared shape of the substitution
// fuzzers.
func eachReplacement(r []rune, cand func(rune) []rune) []string {
	var out []string
	for i := range r {
		for _, c := range cand(r[i]) {
			out = append(out, withReplaced(r, i, c))
		}
	}
	return out
}

// eachSeparator inserts sep between characters, skipping any position adjacent to
// a blocked character — the shared shape of the hyphenation and subdomain fuzzers.
func eachSeparator(r []rune, sep rune, blocked string) []string {
	var out []string
	for i := 1; i < len(r); i++ {
		if strings.ContainsRune(blocked, r[i-1]) || strings.ContainsRune(blocked, r[i]) {
			continue
		}
		out = append(out, withInserted(r, i, sep))
	}
	return out
}

// fuzzReplacement substitutes each character with each of its keyboard neighbours.
func fuzzReplacement(r []rune, _ Options) []string {
	return eachReplacement(r, keyboardAdjacent)
}

// fuzzInsertion inserts each interior character's keyboard neighbours beside it —
// the fat-finger slip of hitting an adjacent key while typing a character.
func fuzzInsertion(r []rune, _ Options) []string {
	var out []string
	for i := 1; i < len(r)-1; i++ {
		for _, c := range keyboardAdjacent(r[i]) {
			out = append(out, withInserted(r, i, c))   // before r[i]
			out = append(out, withInserted(r, i+1, c)) // after r[i]
		}
	}
	return out
}

// fuzzAddition appends each ASCII letter (google → googlea … googlez).
func fuzzAddition(r []rune, _ Options) []string {
	out := make([]string, 0, 26)
	s := string(r)
	for c := 'a'; c <= 'z'; c++ {
		out = append(out, s+string(c))
	}
	return out
}

// fuzzHyphenation inserts a hyphen between characters, never beside an existing
// hyphen (which would make an invalid label).
func fuzzHyphenation(r []rune, _ Options) []string {
	return eachSeparator(r, '-', "-")
}

// fuzzSubdomain inserts a dot between characters, splitting the label into a
// subdomain — never beside an existing dot or hyphen.
func fuzzSubdomain(r []rune, _ Options) []string {
	return eachSeparator(r, '.', "-.")
}

const vowels = "aeiou"

func isVowel(c rune) bool { return strings.ContainsRune(vowels, c) }

// fuzzVowelSwap replaces each vowel with each other vowel.
func fuzzVowelSwap(r []rune, _ Options) []string {
	var out []string
	for i := range r {
		if !isVowel(r[i]) {
			continue
		}
		for _, v := range vowels {
			if v != r[i] {
				out = append(out, withReplaced(r, i, v))
			}
		}
	}
	return out
}

// fuzzHomoglyph substitutes each character with each of its single-rune Unicode
// confusables (see [homoglyphs]).
func fuzzHomoglyph(r []rune, _ Options) []string {
	return eachReplacement(r, func(c rune) []rune { return homoglyphs[c] })
}

// fuzzLeet substitutes each character with its leetspeak numeral (see [leet]) — the
// "reads as" look-alike (paypal → p4ypal), one position at a time so each variant is
// a single edit and twist detects it at distance 1. Multi-position leet (p4yp4l) is
// the composition of several and falls outside this single-edit fuzzer.
func fuzzLeet(r []rune, _ Options) []string {
	return eachReplacement(r, func(c rune) []rune { return leet[c] })
}

// fuzzBitsquatting flips each bit of each ASCII byte, keeping only flips that land
// on a valid label character — the class of typo caused by a single-bit memory or
// transmission error.
func fuzzBitsquatting(r []rune, _ Options) []string {
	var out []string
	for i := range r {
		c := r[i]
		if c > 127 {
			continue // ASCII only
		}
		ci := int(c)
		for b := 0; b < 7; b++ {
			f := rune(ci ^ (1 << b))
			if isLabelByte(f) {
				out = append(out, withReplaced(r, i, f))
			}
		}
	}
	return out
}

// fuzzTLDSwap appends each caller-supplied TLD (google + {net,org} →
// google.net, google.org). twister ships no TLD list, so this yields nothing
// unless [Options.TLDs] is set.
func fuzzTLDSwap(r []rune, o Options) []string {
	if len(o.TLDs) == 0 {
		return nil
	}
	label := string(r)
	out := make([]string, 0, len(o.TLDs))
	for _, t := range o.TLDs {
		t = strings.ToLower(strings.Trim(strings.TrimSpace(t), "."))
		if t == "" {
			continue
		}
		out = append(out, label+"."+t)
	}
	return out
}

// isLabelByte reports whether c is a valid host-label character: a–z, 0–9, or '-'.
func isLabelByte(c rune) bool {
	return (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
}

// --- rune helpers: each returns a new label string, leaving r untouched ---------

func withDeleted(r []rune, i int) string {
	out := make([]rune, 0, len(r)-1)
	out = append(out, r[:i]...)
	out = append(out, r[i+1:]...)
	return string(out)
}

func withInserted(r []rune, i int, c rune) string {
	out := make([]rune, 0, len(r)+1)
	out = append(out, r[:i]...)
	out = append(out, c)
	out = append(out, r[i:]...)
	return string(out)
}

func withReplaced(r []rune, i int, c rune) string {
	out := make([]rune, len(r))
	copy(out, r)
	out[i] = c
	return string(out)
}

func withSwapped(r []rune, i, j int) string {
	out := make([]rune, len(r))
	copy(out, r)
	out[i], out[j] = out[j], out[i]
	return string(out)
}
