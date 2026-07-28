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
| `Permute(label string) []Variant` | full core set, deduped, seed excluded, sorted by name |
| `PermuteWith(label string, o Options) []Variant` | same, with fuzzer selection and a TLD list |
| `Variant{Name, Fuzzer string}` | one permutation and the technique that produced it |
| `Options{Fuzzers []string, TLDs []string}` | `Fuzzers` empty = all; `TLDs` feeds `tld-swap` only |
| `FuzzerNames []string` | the valid fuzzer names, in run order |

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

`omission` · `repetition` · `transposition` · `replacement` · `insertion` ·
`addition` · `hyphenation` · `subdomain` · `vowel-swap` · `homoglyph` ·
`bitsquatting` · `tld-swap`

The nine edit-based fuzzers each make exactly one edit, so every variant is one
Damerau-Levenshtein step from the seed; `homoglyph` and `bitsquatting` are single
substitutions too. `tld-swap` requires `Options.TLDs` (twister ships no TLD list) —
it yields nothing under `Permute`.

## CLI

```
twister permute [-f a,b,c] [-tld com,net,org] [labels...]
twister version
```

- `-f` — comma-separated fuzzer subset (default all).
- `-tld` — comma-separated TLDs for `tld-swap`.
- Labels are the arguments, or one per line on **stdin** if none are given.
- Output: one line per variant — `label⇥variant⇥fuzzer`.

```sh
twister permute paypal                          # every technique
twister permute -f homoglyph,omission paypal    # just two techniques
twister permute -tld com,net,co paypal          # include tld-swap
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

## Cross-validation with twist

The [diff/](../diff/) nested module round-trips twister through `twist`: every
edit-based variant of a seed must be detected by `twist` as a near-miss at distance
1. It is a separate module so the root stays zero-dependency; run it with both repos
checked out as siblings:

```sh
cd diff && GOWORK=off go test ./...
```
