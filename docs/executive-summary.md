# twister — executive summary

**What it is.** A pure-Go, zero-dependency library that generates the typosquat
permutations of a domain label — the offensive-side complement to `twist`. Given a
name it enumerates the look-alike variants an attacker might register, each tagged
with the technique that produced it.

**Why it exists.** Defending a brand against look-alike domains starts with knowing
what the look-alikes *are*. Full squat-hunting tools both generate those permutations
and then resolve/enrich them over the network. twister carves out just the generation
half as a clean, embeddable Go library: deterministic, network-free, and testable, so
any consumer can wrap the networked stages it needs around a permutation engine it can
trust and unit-test.

**How it fits.** twister and `twist` are a matched pair on the same edit-distance
lineage, pointing opposite ways:

- **twister** *generates* — "what could squat my brand?" — proactive, produces the
  candidate space.
- **twist** *detects* — "is this observed name a near-miss of a known brand?" —
  reactive, over a target list.

Because every edit-based fuzzer makes exactly one edit, each generated variant is
one edit-distance step from the seed and is detected by `twist` at distance 1 — a
generate → detect round-trip that cross-validates both libraries at once.

**The boundary.** twister does no resolution, registration, scoring, or enrichment —
no DNS, whois, geoip, or network of any kind. That networked half belongs to a
consumer. twister is pure generation: a label in, a candidate set out.

**Shape.** Two public entry points — `Permute(label)` and `PermuteWith(label,
Options)` — returning `[]Variant{Name, Fuzzer, EditCount}`, deduplicated,
seed-excluded, and sorted. Thirteen single-edit fuzzers form the dist-1 core, plus
four opt-in, off-by-default multi-edit extensions (combosquat, aggressive
multi-substitution, multi-homoglyph, homophone) that each tag their variants with an
edit count. Standard library only, small embedded data tables. A thin `twister
permute` CLI wraps it for shell use.
