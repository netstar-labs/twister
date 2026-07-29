# twister — torture-chamber audit findings (v0.2)

Four-dimension adversarial audit (A simpler · B dedup · C correctness+security ·
D doc-drift) over the v0.2 generation-extensions branch, 2026-07. Findings verified
against the code, then repaired in the same branch and re-validated (gofmt / build /
vet / `-race` / staticcheck clean; test + `diff/` twist round-trip green; aggressive
blowup re-measured bounded). `*` = failure reproduced against a copy.

## CONFIRMED — fixed

### C1 · CORRECTNESS — `tld-swap` reported a false `EditCount` of 1 `*`
`tld-swap` lived in the EditCount-1 `registry`, so the dispatcher tagged every TLD
append as a single edit — but `paypal → paypal.com` is a dot plus three runes.
**Fix:** moved `tld-swap` to `extRegistry` and changed `fuzzTLDSwap` to return
`[]rawVariant` with the true cost `1 + len([]rune(tld))`. Now `paypal.com` → 4,
`paypal.computer` → 9. Guarded by a new assertion in `TestCoreIsSingleEdit`.

### C2 · SECURITY/ROBUSTNESS — combosquat emitted invalid labels from dirty words `*`
`Options.Words` was affixed verbatim, so a keyword containing whitespace or a
leading/trailing `-`/`.` produced an unregistrable or malformed label (`paypal- `,
`.paypal`). **Fix:** `fuzzCombosquat` skips any word that is empty, contains
whitespace, or begins/ends with `-`/`.` before affixing.

### A1/B1 · SIMPLIFY+DEDUP — one substring-substitution kernel
`fuzzMultiHomoglyph` and `fuzzHomophone` each carried their own sorted-key iteration
and run-slice comparison. **Fix:** extracted `applySubstitutionTable(r, table)`
(deterministic via `slices.Sorted(maps.Keys(...))`, matches with `slices.Equal`);
both fuzzers are now one line through it. Removed the hand-rolled `sortedStringKeys`
and `runesEqualAt` helpers (and the `sort` import) in favour of `maps`/`slices`.

### A3 · SIMPLIFY — `mergeHomophones` via `maps`
Replaced the manual copy loop with `out := maps.Clone(base); maps.Copy(out, extra)`.

### A5/A6 · SIMPLIFY — `fuzzAggressive` internals
`cands map[int][]rune` became a parallel `posCands [][]rune` (positions are already
dense), and the `capped bool` sentinel was replaced by `forEachCombination`'s
returned `ok` (stop iterating the moment the `MaxAggressiveVariants` bound is hit).

### B2 · DEDUP — vowel-swap through the shared replacement kernel
`const vowels`/`isVowel`/hand-rolled loop replaced by a static `vowelSwaps` table
fed through the same `eachReplacement` kernel the other substitution fuzzers use.

### C3/C4 · DOC — test-comment accuracy
`TestCoreIsSingleEdit` now documents that the dist-1 invariant is scoped to
`Permute`/zero-options (callers who pass `Words`/`MaxEdits` deliberately inject
multi-edit variants) and asserts `tld-swap` carries `EditCount ≥ 2`.
`TestAggressiveCap` now states the cap is **silent** — `MaxAggressiveVariants` is a
public constant to compare against, not a log line — so the test asserts the bound,
not any stderr.

### D · DOC-DRIFT — v0.2 reality across all docs
- Counts: "thirteen single-edit fuzzers" → **twelve** single-edit core + `tld-swap`
  (README layout, architecture subsystems, introduction, executive-summary).
- Tables: "four embedded tables" → **five** (added the `leet` numeral map) in README
  and architecture.
- `Variant`/`Options` shapes: added `EditCount`, `Words`, `MaxEdits`, `Homophones`
  and `ExtendedFuzzerNames` to the userguide API table; `{Name, Fuzzer}` →
  `{Name, Fuzzer, EditCount}` in the README/architecture data-flow.
- Opt-in framing: "each behind an explicit `Options` field" corrected everywhere to
  the true split — combosquat/aggressive are **data-gated** in `FuzzerNames`
  (`Words`/`MaxEdits`), multi-homoglyph/homophone are **name-gated** via
  `ExtendedFuzzerNames` (doc.go, introduction, README, userguide).
- YAGNI list (architecture): removed combosquat and homophone — they shipped in v0.2.
- Aggressive example `g00gl3` → `g00gle` (the 3-edit form is never emitted; the real
  2-edit output is `g00gle`).
- README multi-homoglyph/homophone example seed `paypal` (which emits nothing) →
  `corn` → `com` (rn→m, 2) · `korn` (c→k, 1).
- CLI: documented the new `-words`/`-maxedits` flags and the added `editcount`
  output column (README, userguide, `app/twister` header).

## REFUTED — kept (with reason)

- **`MaxAggressiveVariants` "should log when it caps"** — a library must not write to
  the global logger. The bound is a public constant the caller compares against; the
  cap is deterministic and silent by design. Kept.
- **`osaDistance` / `editDistance` "duplicate metrics, dedup them"** — deliberately
  separate: `osaDistance` is restricted Damerau-Levenshtein (3-row, counts a
  transposition as one edit) and lets `aggressive` drop two substitutions that are
  really a transposition; `editDistance` is plain Levenshtein (2-row). Merging them
  would break the transposition drop. Kept.
- **Aggressive "combinatorial DoS"** `*` — the blowup reproduces in principle but is
  bounded by `MaxAggressiveVariants` (<40ms even at `MaxEdits=1000`); not a vuln. Kept.
- **`diff/` harness "misses multi-edit variants"** — correct and intended: the
  round-trip asserts twist@1 detection, so its `editFuzzers` list deliberately
  excludes `tld-swap` and all four multi-edit extensions (`TestMultiEditNotDetectedAtDistance1`
  pins this). Kept.
