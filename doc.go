// Package twister generates typosquat permutations of a domain label: hand it a
// registrable label and it enumerates the look-alikes a squatter might register —
// omissions, insertions, repetitions, transpositions, keyboard slips, homoglyphs,
// leetspeak digits, bit flips, TLD swaps. twister is the counter to twist: twister *generates* the
// look-alike permutations and twist *detects* them, so the two are a matched pair —
// twist ↔ twister, detect ↔ generate, the same edit-distance space walked in
// opposite directions.
//
// twister is pure generation and nothing else. It takes a label and returns a set
// of [Variant]s, each tagged with the technique that produced it. It does not
// resolve, register, score, or enrich — no DNS, whois, geoip, or network of any
// kind; that networked half is a consumer's job. The
// output is deterministic, deduplicated, and excludes the seed itself, so the same
// label always yields the same sorted set and a caller can weight or filter by
// technique.
//
// It is pure Go, standard library only. The data-backed fuzzers (keyboard
// adjacency, homoglyph confusables) are small embedded tables, not third-party
// dependencies. twister operates on the registrable label only: a caller splits a
// domain into label + eTLD, permutes the label, and — for the tld-swap fuzzer —
// supplies its own TLD list, since twister ships none. Every edit-based fuzzer
// makes exactly one edit, so each such variant is one Damerau-Levenshtein step
// from the seed and is detected as a near-miss by the sibling twist package — a
// clean generate → detect round-trip. (tld-swap, which appends a TLD, is not an
// edit-based fuzzer.)
package twister
