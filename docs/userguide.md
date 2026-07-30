# twister — user guide

## Library

```go
import "github.com/netstar-labs/twister"

// Full core set (every fuzzer except tld-swap, which needs a TLD list):
for _, v := range twister.Permute("paypal") {
    fmt.Println(v.Name, v.Fuzzer)
}

// Select techniques and/or supply TLDs:
vs := twister.PermuteWith("paypal", twister.Options{
    Fuzzers: []string{"homoglyph", "transposition", "tld-swap"},
    TLDs:    []string{"com", "net", "co"},
})
```

### API

| Symbol | Meaning |
|---|---|
| `Permute(label string) []Variant` | tight dist-1 core set, deduped, seed excluded, sorted by name |
| `PermuteWith(label string, o Options) []Variant` | same, with fuzzer selection and the opt-in data (TLDs/Words/MaxEdits/Homophones) |
| `Variant{Name, Fuzzer string; EditCount int}` | one permutation, the technique that produced it, and its edit distance from the seed |
| `Options{Fuzzers, TLDs, Words []string; MaxEdits int; Homophones map[string][]string}` | `Fuzzers` empty = all in `FuzzerNames`; `TLDs`/`Words`/`MaxEdits` data-gate `tld-swap`/`combosquat`/`aggressive`; `Homophones` overrides the homophone table |
| `FuzzerNames []string` | the default-run fuzzer names, in run order |
| `ExtendedFuzzerNames []string` | the unconditional multi-edit fuzzers (`multi-homoglyph`, `homophone`) — run only when named in `Options.Fuzzers` |

**Contract.** The seed is lower-cased and trimmed. Output is deterministic
(same input → same sorted set), deduplicated, and never contains the seed. Each
`Variant` carries the fuzzer that produced it; on a tie the earlier fuzzer in
`FuzzerNames` owns it. Unknown fuzzer names are ignored; an empty/blank label
returns `nil`. twister is network-free and safe for concurrent use (it shares no
mutable state).

**Input.** twister operates on the **registrable label** — the caller splits a
domain into label + eTLD and normalizes (lowercase, IDNA) first, then permutes the
label. Most fuzzers emit a bare label; `subdomain` emits a dotted label and
`tld-swap` emits `label.tld`.

### Fuzzers

**Single-edit core** (every variant is exactly one Damerau-Levenshtein step from the
seed — the tight dist-1 set `Permute` returns):

`omission` · `repetition` · `transposition` · `replacement` · `insertion` ·
`addition` · `hyphenation` · `subdomain` · `vowel-swap` · `homoglyph` · `leet` ·
`bitsquatting`

The nine edit-based fuzzers each make exactly one edit; `homoglyph`, `leet`, and
`bitsquatting` are single substitutions too.

**Multi-edit (opt-in, v0.2)** — each variant carries its true `EditCount`:

- `tld-swap` — appends a TLD; data-gated on `Options.TLDs` (twister ships no list).
- `combosquat` — `brand-word` / `word-brand` / `brandword` / `brand.word`; data-gated
  on `Options.Words`.
- `aggressive` — substitution fuzzers composed at 2..`MaxEdits` positions (`p4yp4l`);
  data-gated on `Options.MaxEdits ≥ 2`, bounded by `MaxAggressiveVariants`.
- `multi-homoglyph` / `homophone` — emit unconditionally from embedded tables, so
  they are **not** in `FuzzerNames`; enable them by naming them in `Options.Fuzzers`
  (see `ExtendedFuzzerNames`). `Options.Homophones` overrides the homophone table.

The three data-gated fuzzers sit in `FuzzerNames` but stay silent until their option
is set, so `Permute` (and any empty-`Fuzzers` call) never emits them.

## CLI

```
twister permute [-f a,b,c] [-tld com,net,org] [-words login,secure] [-maxedits 2] [labels...]
twister version
```

- `-f` — comma-separated fuzzer subset (default all).
- `-tld` — comma-separated TLDs for `tld-swap`.
- `-words` — comma-separated keywords for `combosquat`.
- `-maxedits` — max positions `aggressive` substitutes at once (≥2 turns it on).
- Labels are the arguments, or one per line on **stdin** if none are given.
- Output: one line per variant — `label⇥variant⇥fuzzer⇥editcount`.

```sh
twister permute paypal                          # every default technique
twister permute -f homoglyph,omission paypal    # just two techniques
twister permute -tld com,net,co paypal          # include tld-swap
twister permute -words login,secure -maxedits 2 paypal   # combosquat + aggressive
printf 'paypal\ngoogle\n' | twister permute      # labels on stdin
```

## Build & install

`build/twister` cross-compiles the CLI to `linux/amd64`, stamps the version from
`git describe`, packages a self-contained installer, and (given a host) scp's it over
and runs it. There is no service — twister is a library + CLI — so the installer just
drops the binary (atomically).

```sh
build/twister                 # package under build/install/, no transmit
build/twister user@host       # package, scp, and install over ssh
```

## Cross-validation with snare

The [diff/](../diff/) nested module round-trips twister through `snare`: every
edit-based variant of a seed must be detected by `snare` as a near-miss at distance
1. It is a separate module so the root stays zero-dependency; run it with both repos
checked out as siblings:

```sh
cd diff && GOWORK=off go test ./...
```
