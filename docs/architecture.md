# twister — architecture

twister is one small library: a dispatcher that fans a seed label out across a set
of independent *fuzzers*, collects their output, deduplicates it, drops the seed,
and returns a sorted, technique-tagged set. There is no state, no I/O, and no
network.

## Data flow

```
                     ┌──────────── fuzzers (registry) ────────────┐
 label ─▶ normalize  │ omission  repetition  transposition  …     │
        (lower/trim) │ replacement  insertion  addition           │ each returns
             │       │ hyphenation  subdomain  vowel-swap          │ raw variants
             ▼       │ homoglyph  bitsquatting  tld-swap           │ (may dup / repeat seed)
          []rune ───▶└──────────────────┬─────────────────────────┘
                                        ▼
                       dedup (seed pre-seeded) + first-fuzzer-owns
                                        ▼
                            sort by Name ─▶ []Variant{Name, Fuzzer}
```

The seed is lower-cased and trimmed, then decoded to `[]rune` once and shared
(read-only) with every fuzzer, so Unicode is handled a rune at a time rather than a
byte at a time. Each fuzzer is a `func([]rune, Options) []string` returning raw
candidate labels; the dispatcher owns correctness of the *set* (dedup, seed
exclusion, ordering) so the fuzzers stay trivially simple.

## Subsystems

| Piece | Responsibility |
|---|---|
| `twister.go` | The public API (`Variant`, `Options`, `Permute`, `PermuteWith`), the fuzzer `registry`, `FuzzerNames`, and the dedup/sort dispatch. |
| `fuzzers.go` | The twelve fuzzers and the four rune-edit helpers (`withDeleted`/`withInserted`/`withReplaced`/`withSwapped`), each returning a fresh string. |
| `tables.go` | The two embedded data tables: QWERTY key adjacency and the homoglyph confusables map. |

## The fuzzers

**Edit-based** — each makes exactly one Damerau-Levenshtein edit, so the variant is
distance 1 from the seed:

| Fuzzer | Edit | Example (`google`) |
|---|---|---|
| `omission` | delete a char | `oogle`, `gogle` |
| `repetition` | double a char | `ggoogle`, `gooogle` |
| `transposition` | swap adjacent chars | `ogogle`, `goolge` |
| `replacement` | substitute a keyboard-adjacent key | `hoogle`, `foogle` |
| `insertion` | insert a keyboard-adjacent key beside an interior char | `gooogle`, `gioogle` |
| `addition` | append a–z | `googlea`…`googlez` |
| `hyphenation` | insert a hyphen | `g-oogle`, `goo-gle` |
| `subdomain` | insert a dot | `g.oogle`, `goo.gle` |
| `vowel-swap` | swap a vowel for another | `gaogle`, `giogle` |

**Data-backed** — small embedded tables, still single-edit where noted:

| Fuzzer | Source | Edit |
|---|---|---|
| `homoglyph` | curated single-rune UTS-39 confusables map | one rune substitution |
| `bitsquatting` | flip each bit of each ASCII byte, keep valid label chars | one substitution |
| `tld-swap` | caller-supplied TLD list (`Options.TLDs`; twister ships none) | append `.tld` (not single-edit) |

`keyboard` in the scope's list is not a standalone fuzzer — it is the QWERTY
adjacency *table* that powers `insertion` and `replacement`.

## Design choices & trade-offs

- **Dispatcher owns the set, fuzzers own the edits.** Dedup, seed exclusion, and
  deterministic ordering live once in `PermuteWith`; a fuzzer just yields raw
  candidates. Adding a fuzzer is a function plus a registry line.
- **First-fuzzer-owns on a tie.** When two techniques produce the same label, the
  earlier one in `FuzzerNames` keeps it. Deterministic, and it attributes a variant
  to its most canonical cause (e.g. `googlel`… a doubled `l` is `repetition`, not
  `addition`).
- **Runes, not bytes.** Homoglyphs emit non-ASCII, and edit distance is defined over
  runes, so the whole engine works on `[]rune`. That keeps every homoglyph a *single*
  edit and matches how `twist` counts.
- **Output is linear per fuzzer.** Each fuzzer emits O(len × small-constant)
  candidates; there is no combinatorial blow-up. Generation is microsecond-class.

## Deliberately out (YAGNI)

- **No resolution / enrichment** — no DNS, whois, geoip, MX, banners, ports. That is
  dnstwist's networked half and a consumer's job; twister never touches the network.
- **No dictionary / keyword / combosquat** (`paypal-secure`) — a different signal;
  add an optional fuzzer later if a wordlist is supplied.
- **No homophone / plural / common-misspelling** dictionaries — data-heavy; deferred.
- **No punycode/IDN encoding** — homoglyphs emit Unicode; leave punycode to the
  caller.
- **No scoring / ranking** — twister enumerates; weighting is the consumer's, or a
  `twist` round-trip (generate, then score each variant against the brand).
