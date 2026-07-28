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
             ▼       │ homoglyph  leet  bitsquatting  tld-swap     │ (may dup / repeat seed)
          []rune ───▶└──────────────────┬─────────────────────────┘
                                        ▼
                       dedup (seed pre-seeded) + first-fuzzer-owns
                                        ▼
                            sort by Name ─▶ []Variant{Name, Fuzzer, EditCount}
```

The seed is lower-cased and trimmed, then decoded to `[]rune` once and shared
(read-only) with every fuzzer, so Unicode is handled a rune at a time rather than a
byte at a time. Each fuzzer is a `func([]rune, Options) []string` returning raw
candidate labels; the dispatcher owns correctness of the *set* (dedup, seed
exclusion, ordering) so the fuzzers stay trivially simple.

## Subsystems

| Piece | Responsibility |
|---|---|
| `twister.go` | The public API (`Variant`, `Options`, `Permute`, `PermuteWith`), the fuzzer `registry`/`extRegistry`, `FuzzerNames`/`ExtendedFuzzerNames`, and the dedup/sort dispatch. |
| `fuzzers.go` | The thirteen single-edit core fuzzers and the four rune-edit helpers (`withDeleted`/`withInserted`/`withReplaced`/`withSwapped`), each returning a fresh string. |
| `extended.go` | The four opt-in **multi-edit** fuzzers (`combosquat`, `aggressive`, `multi-homoglyph`, `homophone`), their `extRegistry`, and their helpers (combination/product enumeration, `osaDistance`). |
| `tables.go` | The four embedded data tables: QWERTY key adjacency, the single-rune homoglyph confusables map, the multi-rune `multiHomoglyphs` sequence map, and the `homophones` sound-alike map. |

## The fuzzers

**Edit-based** — each makes exactly one Damerau-Levenshtein edit, so the variant is
distance 1 from the seed:

| Fuzzer | Edit | Example (`google`) |
|---|---|---|
| `omission` | delete a char | `oogle`, `gogle` |
| `repetition` | double a char | `ggoogle`, `gooogle` |
| `transposition` | swap adjacent chars | `ogogle`, `goolge` |
| `replacement` | substitute a keyboard-adjacent key | `hoogle`, `foogle` |
| `insertion` | insert a keyboard-adjacent key beside an interior char | `gioogle`, `g0oogle` |
| `addition` | append a–z | `googlea`…`googlez` |
| `hyphenation` | insert a hyphen | `g-oogle`, `goo-gle` |
| `subdomain` | insert a dot | `g.oogle`, `goo.gle` |
| `vowel-swap` | swap a vowel for another | `gaogle`, `giogle` |

**Data-backed** — small embedded tables, still single-edit where noted:

| Fuzzer | Source | Edit |
|---|---|---|
| `homoglyph` | curated single-rune UTS-39 confusables map | one rune substitution |
| `leet` | leetspeak numeral map (`a→4`, `e→3`, `o→0`, …) | one rune substitution |
| `bitsquatting` | flip each bit of each ASCII byte, keep valid label chars | one substitution |
| `tld-swap` | caller-supplied TLD list (`Options.TLDs`; twister ships none) | append `.tld` (not single-edit) |

`keyboard` in the scope's list is not a standalone fuzzer — it is the QWERTY
adjacency *table* that powers `insertion` and `replacement`.

## The multi-edit extensions (v0.2, opt-in)

The core above is deliberately single-edit, and `Permute` stays that tight dist-1
set. Four additional fuzzers (in `extended.go`) generate **multi-edit** look-alikes.
Every one is OFF by default and each variant carries its true `EditCount` (the
Damerau-Levenshtein distance from the seed), so a consumer routes it to the detector
that can recover it rather than expecting a twist@1 hit.

| Fuzzer | Turned on by | Emits | Edit count |
|---|---|---|---|
| `combosquat` | `Options.Words` (keywords) | `brand-word`, `word-brand`, `brandword`, `wordbrand`, `brand.word` | runes added (>1) |
| `aggressive` | `Options.MaxEdits ≥ 2` | substitution fuzzers composed at 2..MaxEdits positions (`p4yp4l`, `g00gl3`) | true OSA distance (2+) |
| `multi-homoglyph` | named in `Options.Fuzzers` | `multiHomoglyphs` rules: `rn→m`, `vv→w`, `cl→d`, `nn→m`, `m→rn` | 2 |
| `homophone` | named in `Options.Fuzzers` (`Options.Homophones` overrides the table) | sound-alike substrings: `ph↔f`, `c↔k`, `s↔z`, `ck↔k` | edit cost of the rewrite |

`combosquat` and `aggressive` are **data-gated** — they sit in `FuzzerNames` but stay
silent until their `Options` field is set (exactly like `tld-swap`), so an
empty-`Fuzzers` call including `Permute` never emits them. `multi-homoglyph` and
`homophone` emit from embedded tables unconditionally, so they are kept OUT of
`FuzzerNames` (they live in `ExtendedFuzzerNames`) and run only when named — otherwise
they would break `Permute`'s dist-1 invariant.

`aggressive` is the one combinatorial fuzzer: it enumerates every k-combination of
substitutable positions × the product of per-position candidates. That explodes with
`MaxEdits`, so its output is bounded by the exported `MaxAggressiveVariants` constant —
a deterministic cap the caller can compare against, not a global-logger side effect. It
tags each variant with `osaDistance(seed, variant)`
and drops any that collapse to a single transposition (distance 1) — those belong to
the core `transposition` fuzzer, not here.

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
- **Output is linear per core fuzzer.** Each single-edit fuzzer emits O(len ×
  small-constant) candidates; there is no combinatorial blow-up, and generation is
  microsecond-class. The one exception is the opt-in `aggressive` fuzzer, whose
  candidate space is combinatorial by design and therefore explicitly bounded by the
  exported `MaxAggressiveVariants` constant (a deterministic cap, not a logged one).

## Deliberately out (YAGNI)

- **No resolution / enrichment** — no DNS, whois, geoip, MX, banners, ports. That is
  the networked half of squat-hunting and a consumer's job; twister never touches the
  network.
- **No dictionary / keyword / combosquat** (`paypal-secure`) — a different signal;
  add an optional fuzzer later if a wordlist is supplied.
- **No homophone / plural / common-misspelling** dictionaries — data-heavy; deferred.
- **No punycode/IDN encoding** — homoglyphs emit Unicode; leave punycode to the
  caller.
- **No scoring / ranking** — twister enumerates; weighting is the consumer's, or a
  `twist` round-trip (generate, then score each variant against the brand).
