# twister

Typosquat **permutation generation** for Go — **hand it a label and it enumerates
the look-alikes a squatter might register**, pure standard library, no
dependencies. It is the counter to the sibling `twist`: where `twist` *detects* a
near-miss, twister *generates* the candidates. twist ↔ twister, detect ↔ generate —
the same edit-distance space, walked in opposite directions.

```
seed label ─▶ PermuteWith ─▶ fan out over fuzzers ─▶ dedup + drop seed ─▶ sort ─▶ []Variant
                              │                                                    │
   omission · repetition · transposition · replacement · insertion · addition     ▼
   hyphenation · subdomain · vowel-swap · homoglyph · leet · bitsquatting · tld-swap  {Name, Fuzzer}
```

Every edit-based fuzzer makes exactly one edit, so each variant is one
Damerau-Levenshtein step from the seed — and `twist` detects it as a near-miss, a
clean generate → detect round-trip (see [diff/](diff/)).

An opt-in v0.2 layer adds four **multi-edit** fuzzers — combosquat
(`paypal-login`), aggressive multi-substitution (`p4yp4l`), multi-character
homoglyphs (`rn`→`m`), and homophones (`ph`↔`f`). Each is off by default behind an
explicit `Options` field, and every `Variant` is tagged with its `EditCount` so the
tight dist-1 core `Permute` returns stays exactly that.

```go
// combosquat and aggressive are gated on their Options field, like tld-swap:
twister.PermuteWith("paypal", twister.Options{
    Words:    []string{"login", "secure"},   // → paypal-login, secure-paypal, paypallogin, …
    MaxEdits: 2,                              // → p4yp4l, … (aggressive multi-substitution)
})
// multi-homoglyph and homophone emit unconditionally, so name them explicitly:
twister.PermuteWith("paypal", twister.Options{Fuzzers: []string{"multi-homoglyph", "homophone"}})
```

## Quick start

```go
for _, v := range twister.Permute("paypal") {
    fmt.Println(v.Name, v.Fuzzer)   // sorted by name: papyal transposition · paypa1 homoglyph · paypall repetition · …
}

// select techniques and supply a TLD list for tld-swap (twister ships none):
twister.PermuteWith("paypal", twister.Options{
    Fuzzers: []string{"homoglyph", "tld-swap"},
    TLDs:    []string{"com", "net", "co"},
})   // paypal.com, paypal.net, paypal.co, pаypal (Cyrillic a), …
```

```sh
go run ./app/twister permute paypal                     # variant per line: label⇥variant⇥fuzzer
go run ./app/twister permute -f homoglyph,omission paypal
printf 'paypal\ngoogle\n' | go run ./app/twister permute -tld com,net
```

## Documentation

- **Start here** — [docs/introduction.md](docs/introduction.md) ·
  [docs/executive-summary.md](docs/executive-summary.md)
- **Deep dive** — [docs/architecture.md](docs/architecture.md)
- **Operations** — [docs/userguide.md](docs/userguide.md)
- **Examples** — [example/README.md](example/README.md)

## Layout

| File | Purpose |
|---|---|
| [twister.go](twister.go) | `Variant`, `Options`, `Permute`/`PermuteWith`, the fuzzer registries and dedup/sort dispatch, `FuzzerNames`/`ExtendedFuzzerNames` |
| [fuzzers.go](fuzzers.go) | the thirteen single-edit core fuzzers (edit-based + data-backed) and the rune-edit helpers |
| [extended.go](extended.go) | the four opt-in multi-edit fuzzers (combosquat, aggressive, multi-homoglyph, homophone) and their helpers |
| [tables.go](tables.go) | the embedded QWERTY adjacency, homoglyph, multi-homoglyph, and homophone tables |
| [doc.go](doc.go) | package doc — the name metaphor (generates, not detects) and the pure-generation scope |
| [app/twister/](app/twister/main.go) | the CLI — `permute` · `version` |
| [diff/](diff/README.md) | the twister ↔ twist differential harness (nested module, keeps the root zero-dep) |

## Notes

- Go module `github.com/netstar-labs/twister`. **Standard library only** — no
  dependencies. Build standalone with `GOWORK=off`. The data-backed fuzzers use
  small embedded tables, not third-party packages.
- twister operates on the **registrable label**: the caller splits a domain into
  label + eTLD (normalize / lowercase / IDNA first), permutes the label, and
  supplies its own TLD list for `tld-swap`.
- **Pure generation, network-free.** No resolution, registration, scoring, or
  enrichment — that networked half is a consumer's job; see
  [docs/architecture.md](docs/architecture.md) § "Deliberately out (YAGNI)".
