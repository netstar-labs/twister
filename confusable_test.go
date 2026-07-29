package twister

import "testing"

func TestConfusableClass(t *testing.T) {
	cases := []struct {
		a, b rune
		want string
	}{
		{'a', '4', "leet"},      // leetspeak numeral
		{'e', '3', "leet"},      //
		{'o', '0', "homoglyph"}, // 0 is in both o's homoglyph and leet sets — homoglyph wins
		{'l', '1', "homoglyph"}, // curated single-rune confusable
		{'a', 'à', "homoglyph"}, // Latin-1 diacritic
		{'a', 'а', "homoglyph"}, // Cyrillic a
		{'s', 'd', "keyboard"},  // QWERTY neighbours (s: edxzaw)
		{'a', 'z', "keyboard"},  // z is adjacent to a (a: qwsz)
		{'a', 'm', ""},          // not a look-alike of a
		{'a', 'a', ""},          // identical is never confusable
	}
	for _, c := range cases {
		if got := ConfusableClass(c.a, c.b); got != c.want {
			t.Errorf("ConfusableClass(%q, %q) = %q, want %q", c.a, c.b, got, c.want)
		}
		if got, want := Confusable(c.a, c.b), c.want != ""; got != want {
			t.Errorf("Confusable(%q, %q) = %v, want %v", c.a, c.b, got, want)
		}
	}
}

// TestConfusableFeedsWeightedCost documents the intended use: the class maps to a
// substitution cost so a caller can build twist.NearestWeighted's Sub without importing
// twister's tables. Two leet/homoglyph edits stay cheap; a random edit is full price.
func TestConfusableFeedsWeightedCost(t *testing.T) {
	sub := func(a, b rune) float64 {
		switch ConfusableClass(a, b) {
		case "homoglyph":
			return 0.25
		case "leet":
			return 0.35
		case "keyboard":
			return 0.60
		default:
			return 1.0
		}
	}
	// paypal -> p4yp4l is two leet swaps: 0.35 + 0.35 = 0.70, within a ~1.0 budget.
	if c := sub('a', '4') + sub('a', '4'); c >= 1.0 {
		t.Errorf("two leet swaps cost %.2f, want < 1.0 (should be admissible)", c)
	}
	// A random swap alone already costs a full edit.
	if c := sub('a', 'm'); c != 1.0 {
		t.Errorf("random swap cost %.2f, want 1.0", c)
	}
}
