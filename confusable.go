package twister

import "slices"

// Confusable reports whether b is a known look-alike substitution for a — a homoglyph,
// a leetspeak numeral, or a QWERTY-adjacent key. It is [ConfusableClass] reduced to a
// yes/no; a == b is never confusable.
func Confusable(a, b rune) bool {
	return ConfusableClass(a, b) != ""
}

// ConfusableClass reports the class by which b is a look-alike substitution for a:
// "homoglyph" (a visual UTS-39 confusable — à or Cyrillic а for a), "leet" (a leetspeak
// numeral — 4 for a), or "keyboard" (a QWERTY-adjacent key — s for a). It returns "" when
// b is not a known look-alike for a, or when a == b. When several classes apply (0 is
// both a homoglyph and the leet form of o) the strongest visual signal wins, in the
// order homoglyph, leet, keyboard.
//
// It exposes the same per-rune substitution data the aggressive fuzzer composes
// internally ([substitutionCandidates]), so a consumer can build a confusability-weighted
// substitution cost for snare.NearestWeighted — a homoglyph or leet swap priced cheap, a
// random swap full price — without re-curating the tables. snare stays zero-dependency by
// taking that cost function from its caller; this is the data that function reads.
//
// Inputs are matched against the tables' lowercase-ASCII keys, so normalize (lower-case)
// the runes first, exactly as snare and twister callers already do.
func ConfusableClass(a, b rune) string {
	if a == b {
		return ""
	}
	switch {
	case slices.Contains(homoglyphs[a], b):
		return "homoglyph"
	case slices.Contains(leet[a], b):
		return "leet"
	case slices.Contains(keyboard[a], b):
		return "keyboard"
	}
	return ""
}
