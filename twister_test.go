package twister

import (
	"reflect"
	"sort"
	"strings"
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

func TestLeet(t *testing.T) {
	// b->8, a->4, t->7 — one leet substitution each, and none of b/a/t is a
	// homoglyph, so the leet fuzzer alone yields exactly these three.
	got := namesFor("bat", "leet")
	want := []string{"8at", "b4t", "ba7"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("leet(bat) = %v\nwant %v", got, want)
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

func TestNoBoundaryHyphenOrDot(t *testing.T) {
	// Regression for the audit's F1: bitsquatting can flip an ASCII byte to '-',
	// and omission/transposition can expose a boundary hyphen on a seed that
	// already contains one — all produce unregistrable labels that must be dropped.
	for _, seed := range []string{"ibm", "team", "x-ray", "e-shop", "a-b-c", "mmm"} {
		for _, v := range PermuteWith(seed, Options{TLDs: []string{"com"}}) {
			if strings.HasPrefix(v.Name, "-") || strings.HasSuffix(v.Name, "-") ||
				strings.HasPrefix(v.Name, ".") || strings.HasSuffix(v.Name, ".") {
				t.Errorf("seed %q: emitted boundary-invalid label %q (%s)", seed, v.Name, v.Fuzzer)
			}
		}
	}
	// The specific reproductions from the audit must be gone.
	for _, v := range Permute("ibm") {
		if v.Name == "ib-" {
			t.Error("ib- still emitted for ibm")
		}
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

// editCounts returns a name→EditCount map for a single fuzzer over label.
func editCounts(label string, o Options) map[string]int {
	m := map[string]int{}
	for _, v := range PermuteWith(label, o) {
		m[v.Name] = v.EditCount
	}
	return m
}

// --- v0.2 multi-edit extensions ------------------------------------------------

// TestCoreIsSingleEdit is the hard-rule guard: Permute (and any empty-Fuzzers call)
// is the tight dist-1 set — every variant is exactly one edit, no extension leaks in.
func TestCoreIsSingleEdit(t *testing.T) {
	for _, seed := range []string{"paypal", "google", "amazon", "verylongbrandname"} {
		// With only TLDs supplied, the data-gated tld-swap appears (carrying its true
		// multi-edit EditCount), but the unconditional extensions never do, and every
		// non-tld-swap variant is a single edit. NOTE the dist-1 guarantee is scoped to
		// Permute / zero-options — supplying Words or MaxEdits deliberately injects
		// combosquat/aggressive multi-edit variants (asserted in their own tests).
		for _, v := range PermuteWith(seed, Options{TLDs: []string{"com"}}) {
			if v.Fuzzer == "multi-homoglyph" || v.Fuzzer == "homophone" {
				t.Errorf("%s: extension %q ran under empty Fuzzers", seed, v.Fuzzer)
			}
			// tld-swap appends a TLD (a genuine multi-edit); everything else must be 1.
			if v.Fuzzer != "tld-swap" && v.EditCount != 1 {
				t.Errorf("%s: %q (%s) has EditCount %d, want 1 in the core set", seed, v.Name, v.Fuzzer, v.EditCount)
			}
			if v.Fuzzer == "tld-swap" && v.EditCount < 2 {
				t.Errorf("%s: tld-swap %q has EditCount %d, want its true (>1) distance", seed, v.Name, v.EditCount)
			}
		}
	}
	// Plain Permute: strictly dist-1, no tld-swap either.
	for _, v := range Permute("paypal") {
		if v.EditCount != 1 {
			t.Errorf("Permute variant %q (%s) has EditCount %d, want 1", v.Name, v.Fuzzer, v.EditCount)
		}
	}
}

// TestExtensionsOffByDefault confirms every new mode is silent until its Options
// field (or explicit name) turns it on, mirroring tld-swap.
func TestExtensionsOffByDefault(t *testing.T) {
	for _, v := range Permute("paypal") {
		switch v.Fuzzer {
		case "combosquat", "aggressive", "multi-homoglyph", "homophone":
			t.Errorf("extension %q ran under Permute: %q", v.Fuzzer, v.Name)
		}
	}
	// combosquat named but no Words → nothing (like tld-swap named with no TLDs).
	if got := PermuteWith("paypal", Options{Fuzzers: []string{"combosquat"}}); got != nil {
		t.Errorf("combosquat without Words produced %v", got)
	}
	// aggressive named but MaxEdits unset (or < 2) → nothing.
	if got := PermuteWith("paypal", Options{Fuzzers: []string{"aggressive"}}); got != nil {
		t.Errorf("aggressive without MaxEdits produced %v", got)
	}
	if got := PermuteWith("paypal", Options{Fuzzers: []string{"aggressive"}, MaxEdits: 1}); got != nil {
		t.Errorf("aggressive with MaxEdits=1 produced %v", got)
	}
}

func TestCombosquat(t *testing.T) {
	o := Options{Fuzzers: []string{"combosquat"}, Words: []string{"login"}}
	got := namesForOpts("paypal", o)
	want := []string{"login-paypal", "loginpaypal", "paypal-login", "paypal.login", "paypallogin"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("combosquat(paypal, login) = %v\nwant %v", got, want)
	}
	// Every combosquat variant adds runes, so EditCount > 1 for a real keyword.
	for name, ec := range editCounts("paypal", o) {
		if ec <= 1 {
			t.Errorf("combosquat %q EditCount %d, want > 1", name, ec)
		}
	}
	// Blank/whitespace words are skipped, not emitted as bare-brand affixes.
	if got := PermuteWith("paypal", Options{Fuzzers: []string{"combosquat"}, Words: []string{"", "  "}}); got != nil {
		t.Errorf("combosquat with only blank words produced %v", got)
	}
}

func TestMultiHomoglyph(t *testing.T) {
	cases := []struct {
		seed string
		want []string
	}{
		{"corn", []string{"com"}},        // rn→m
		{"com", []string{"corn"}},        // m→rn (reverse)
		{"vvcl", []string{"vvd", "wcl"}}, // cl→d and vv→w
	}
	for _, c := range cases {
		got := namesFor(c.seed, "multi-homoglyph")
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("multi-homoglyph(%q) = %v\nwant %v", c.seed, got, c.want)
		}
		for name, ec := range editCounts(c.seed, Options{Fuzzers: []string{"multi-homoglyph"}}) {
			if ec != 2 {
				t.Errorf("multi-homoglyph %q EditCount %d, want 2", name, ec)
			}
		}
	}
}

func TestHomophone(t *testing.T) {
	// ph→f is a two-edit rewrite; c↔k and s↔z are one edit — each tagged truthfully.
	if got := namesFor("phone", "homophone"); !reflect.DeepEqual(got, []string{"fone"}) {
		t.Errorf("homophone(phone) = %v, want [fone]", got)
	}
	if got := editCounts("phone", Options{Fuzzers: []string{"homophone"}})["fone"]; got != 2 {
		t.Errorf("homophone fone EditCount = %d, want 2 (ph→f)", got)
	}
	got := namesFor("cats", "homophone")
	if !reflect.DeepEqual(got, []string{"catz", "kats"}) {
		t.Errorf("homophone(cats) = %v, want [catz kats]", got)
	}
	for name, ec := range editCounts("cats", Options{Fuzzers: []string{"homophone"}}) {
		if ec != 1 {
			t.Errorf("homophone %q EditCount %d, want 1", name, ec)
		}
	}
	// Options.Homophones overrides the embedded table per key.
	o := Options{Fuzzers: []string{"homophone"}, Homophones: map[string][]string{"c": {"s"}}}
	if got := namesForOpts("cat", o); !reflect.DeepEqual(got, []string{"sat"}) {
		t.Errorf("homophone(cat) with c→s override = %v, want [sat]", got)
	}
}

func TestAggressive(t *testing.T) {
	o := Options{Fuzzers: []string{"aggressive"}, MaxEdits: 2}
	ec := editCounts("paypal", o)
	if _, ok := ec["p4yp4l"]; !ok {
		t.Error("aggressive missing the canonical 2-leet variant p4yp4l")
	}
	if len(ec) == 0 {
		t.Fatal("aggressive produced nothing for paypal MaxEdits=2")
	}
	for name, n := range ec {
		if n != 2 {
			t.Errorf("aggressive %q EditCount %d, want 2 at MaxEdits=2", name, n)
		}
	}
	// MaxEdits=3 tags variants with their real position count (2 or 3).
	for _, v := range PermuteWith("cloudflare", Options{Fuzzers: []string{"aggressive"}, MaxEdits: 3}) {
		if v.EditCount < 2 || v.EditCount > 3 {
			t.Errorf("aggressive(MaxEdits=3) %q EditCount %d, want 2..3", v.Name, v.EditCount)
			break
		}
	}
}

// TestAggressiveCap forces the combinatorial cap and asserts the output stays
// bounded and deterministic. The cap is silent — MaxAggressiveVariants is a public
// constant, not a log line (a library must not write to the global logger) — so the
// test asserts the bound, not any stderr output.
func TestAggressiveCap(t *testing.T) {
	o := Options{Fuzzers: []string{"aggressive"}, MaxEdits: 3}
	a := PermuteWith("verylongbrandname", o)
	if len(a) == 0 {
		t.Fatal("aggressive produced nothing")
	}
	if len(a) > MaxAggressiveVariants {
		t.Errorf("aggressive output %d exceeds cap %d", len(a), MaxAggressiveVariants)
	}
	if b := PermuteWith("verylongbrandname", o); !reflect.DeepEqual(a, b) {
		t.Error("aggressive is not deterministic under the cap")
	}
}

// TestExtensionHygiene runs each extension over a seed that exercises it and checks
// the shared invariants: deterministic, deduped, correctly tagged, sorted.
func TestExtensionHygiene(t *testing.T) {
	cases := []struct {
		fuzzer string
		seed   string
		o      Options
	}{
		{"combosquat", "paypal", Options{Fuzzers: []string{"combosquat"}, Words: []string{"login", "secure"}}},
		{"aggressive", "paypal", Options{Fuzzers: []string{"aggressive"}, MaxEdits: 2}},
		{"multi-homoglyph", "communication", Options{Fuzzers: []string{"multi-homoglyph"}}},
		{"homophone", "processing", Options{Fuzzers: []string{"homophone"}}},
	}
	for _, c := range cases {
		a := PermuteWith(c.seed, c.o)
		if len(a) == 0 {
			t.Errorf("%s produced nothing for %q", c.fuzzer, c.seed)
			continue
		}
		if b := PermuteWith(c.seed, c.o); !reflect.DeepEqual(a, b) {
			t.Errorf("%s is not deterministic", c.fuzzer)
		}
		seen := map[string]bool{}
		for i, v := range a {
			if v.Fuzzer != c.fuzzer {
				t.Errorf("%s: variant %q tagged %q", c.fuzzer, v.Name, v.Fuzzer)
			}
			if v.EditCount < 1 {
				t.Errorf("%s: variant %q untagged EditCount %d", c.fuzzer, v.Name, v.EditCount)
			}
			if v.Name == c.seed {
				t.Errorf("%s: seed not excluded", c.fuzzer)
			}
			if seen[v.Name] {
				t.Errorf("%s: duplicate variant %q", c.fuzzer, v.Name)
			}
			seen[v.Name] = true
			if i > 0 && a[i-1].Name > v.Name {
				t.Errorf("%s: not sorted: %q > %q", c.fuzzer, a[i-1].Name, v.Name)
			}
		}
	}
}
