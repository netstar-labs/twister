package twister

import (
	"reflect"
	"sort"
	"testing"
)

// namesFor returns the sorted variant labels a single fuzzer produces for label.
func namesFor(label, fuzzer string) []string {
	return namesForOpts(label, Options{Fuzzers: []string{fuzzer}})
}

func namesForOpts(label string, o Options) []string {
	var out []string
	for _, v := range PermuteWith(label, o) {
		out = append(out, v.Name)
	}
	sort.Strings(out)
	return out
}

func TestOmission(t *testing.T) {
	got := namesFor("google", "omission")
	want := []string{"gogle", "googe", "googl", "goole", "oogle"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("omission(google) = %v\nwant %v", got, want)
	}
}

func TestTransposition(t *testing.T) {
	got := namesFor("google", "transposition")
	want := []string{"gogole", "googel", "goolge", "ogogle"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("transposition(google) = %v\nwant %v", got, want)
	}
}

func TestRepetition(t *testing.T) {
	got := namesFor("abc", "repetition")
	want := []string{"aabc", "abbc", "abcc"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("repetition(abc) = %v\nwant %v", got, want)
	}
}

func TestHyphenation(t *testing.T) {
	got := namesFor("abc", "hyphenation")
	want := []string{"a-bc", "ab-c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("hyphenation(abc) = %v\nwant %v", got, want)
	}
}

func TestSubdomain(t *testing.T) {
	got := namesFor("abc", "subdomain")
	want := []string{"a.bc", "ab.c"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("subdomain(abc) = %v\nwant %v", got, want)
	}
}

func TestAddition(t *testing.T) {
	got := namesFor("go", "addition")
	if len(got) != 26 {
		t.Fatalf("addition(go) produced %d variants, want 26", len(got))
	}
	if got[0] != "goa" || got[25] != "goz" {
		t.Errorf("addition bounds = %q..%q, want goa..goz", got[0], got[25])
	}
}

func TestVowelSwap(t *testing.T) {
	got := namesFor("bat", "vowel-swap")
	want := []string{"bet", "bit", "bot", "but"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("vowel-swap(bat) = %v\nwant %v", got, want)
	}
}

func TestReplacementKeyboard(t *testing.T) {
	// 'a' neighbours are q,w,s,z — replacing the sole 'a' yields exactly those.
	got := namesFor("a", "replacement")
	want := []string{"q", "s", "w", "z"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("replacement(a) = %v\nwant %v", got, want)
	}
}

func TestHomoglyph(t *testing.T) {
	set := map[string]bool{}
	for _, v := range PermuteWith("paypal", Options{Fuzzers: []string{"homoglyph"}}) {
		if v.Fuzzer != "homoglyph" {
			t.Errorf("wrong tag %q", v.Fuzzer)
		}
		set[v.Name] = true
	}
	if !set["paypa1"] { // l -> 1 is a curated confusable
		t.Error("expected homoglyph paypa1 (l->1)")
	}
	if !set["pаypal"] { // second 'a' -> Cyrillic а
		t.Error("expected homoglyph with Cyrillic a")
	}
}

func TestBitsquattingValidLabels(t *testing.T) {
	got := PermuteWith("amazon", Options{Fuzzers: []string{"bitsquatting"}})
	if len(got) == 0 {
		t.Fatal("bitsquatting produced nothing")
	}
	for _, v := range got {
		for _, c := range v.Name {
			if !isLabelByte(c) {
				t.Errorf("bitsquat produced invalid label char %q in %q", c, v.Name)
			}
		}
	}
}

func TestTLDSwap(t *testing.T) {
	got := namesForOpts("google", Options{Fuzzers: []string{"tld-swap"}, TLDs: []string{"net", "org", ".com", ""}})
	want := []string{"google.com", "google.net", "google.org"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("tld-swap = %v\nwant %v", got, want)
	}
	// Permute (no TLDs) must not emit tld-swap variants.
	for _, v := range Permute("google") {
		if v.Fuzzer == "tld-swap" {
			t.Errorf("tld-swap ran without TLDs: %q", v.Name)
		}
	}
}

func TestHygiene(t *testing.T) {
	a := Permute("paypal")
	if b := Permute("paypal"); !reflect.DeepEqual(a, b) {
		t.Fatal("Permute is not deterministic")
	}
	if len(a) == 0 {
		t.Fatal("Permute(paypal) produced nothing")
	}
	seen := map[string]bool{}
	for i, v := range a {
		if v.Name == "paypal" {
			t.Error("seed not excluded")
		}
		if v.Name == "" {
			t.Error("empty variant name")
		}
		if seen[v.Name] {
			t.Errorf("duplicate variant %q", v.Name)
		}
		seen[v.Name] = true
		if v.Fuzzer == "" {
			t.Errorf("untagged variant %q", v.Name)
		}
		if i > 0 && a[i-1].Name > v.Name {
			t.Errorf("not sorted: %q > %q", a[i-1].Name, v.Name)
		}
	}
}

func TestEmptyAndShort(t *testing.T) {
	if v := Permute(""); v != nil {
		t.Errorf("Permute(\"\") = %v, want nil", v)
	}
	if v := Permute("   "); v != nil {
		t.Errorf("Permute(blank) = %v, want nil", v)
	}
	// A single char still generates (addition, replacement, homoglyph …) but never panics.
	_ = Permute("a")
}

func TestUnknownFuzzerIgnored(t *testing.T) {
	if v := PermuteWith("google", Options{Fuzzers: []string{"nope"}}); v != nil {
		t.Errorf("unknown fuzzer produced %v, want nil", v)
	}
}

func TestNormalizesSeed(t *testing.T) {
	// Upper-case and surrounding space are normalized before permuting.
	a := Permute("Google")
	b := Permute("google")
	if !reflect.DeepEqual(a, b) {
		t.Error("Permute did not normalize case")
	}
}
