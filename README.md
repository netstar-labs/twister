# twister

Typosquat **permutation generation** for Go — **hand it a label and it enumerates
the look-alikes a squatter might register**, pure standard library, no
dependencies. Named after `dnstwist`, and it works the tool's offensive way round:
where the sibling `twist` *detects* a near-miss, twister *generates* the candidates.
twist ↔ twister, detect ↔ generate — the same edit-distance space, walked in
opposite directions.

```
seed label ─▶ PermuteWith ─▶ fan out over fuzzers ─▶ dedup + drop seed ─▶ sort ─▶ []Variant
                              │                                                    │
   omission · repetition · transposition · replacement · insertion · addition     ▼
   hyphenation · subdomain · vowel-swap · homoglyph · bitsquatting · tld-swap   {Name, Fuzzer}
```

Every edit-based fuzzer makes exactly one edit, so each variant is one
Damerau-Levenshtein step from the seed — and `twist` detects it as a near-miss, a
clean generate → detect round-trip (see [diff/](diff/)).

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
| [twister.go](twister.go) | `Variant`, `Options`, `Permute`/`PermuteWith`, the fuzzer registry and dedup/sort dispatch, `FuzzerNames` |
| [fuzzers.go](fuzzers.go) | the twelve fuzzers (edit-based + data-backed) and the rune-edit helpers |
| [tables.go](tables.go) | the embedded QWERTY adjacency and homoglyph confusables tables |
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
