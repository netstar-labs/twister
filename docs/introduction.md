# Meet twister — the permutation, not the near-miss

Every phishing kit begins with a lie told in a domain name. `paypa1.com` for
`paypal.com`; `gooogle`; `arnazon`; `paypaⅼ` with a look-alike Cyrillic character
you would swear was an `l`. Before a defender can catch one of those, someone has to
enumerate the space of what a squatter *would* register — and that is the offensive
half of the canonical tool **dnstwist**: give it a brand and it *generates* the
thousands of plausible permutations. twister is named for that lineage and keeps
exactly that half. Hand it a label and it answers one question: *what could squat
this?*

## What it actually is

twister is a pure-Go, zero-dependency **permutation generator**. You give it a
registrable label — `paypal`, `google`, whatever you are protecting — and it returns
the set of look-alike labels an attacker might register, each tagged with the
technique that produced it. It walks a dozen fuzzers: dropping a character
(`paypa1`→`papal`), doubling one, swapping neighbours (`papyal`), a keyboard slip to
an adjacent key, a hyphen or a dot slipped in, a vowel swapped, a Unicode homoglyph,
a single flipped bit, a different TLD. Each edit-based fuzzer makes *exactly one*
change, so every variant is one edit-distance step from the seed — which is what
makes twister the mirror image of `twist`: what twister generates, twist detects.

## The line it will not cross

dnstwist does more than generate: it then resolves every candidate, runs whois,
geoip, banner grabs, MX and port checks. twister does **none** of that. It is a pure
function — a label in, a candidate set out — with no DNS, no whois, no network of any
kind. That is a deliberate boundary, not a missing feature: resolution and
enrichment are a networked stage a *consumer* owns, and keeping them out is what lets
twister be deterministic, side-effect-free, and unit-tested in isolation. It also
does not score or rank; it enumerates, tags each variant by technique, and leaves the
weighting to the caller (or to a `twist` round-trip that scores each variant back
against the brand).

## The scope it keeps

twister generates the near-space of a name and refuses the neighbouring jobs. It has
no opinion on which variant is most dangerous, no idea whether any candidate is
registered or resolves, and no knowledge of what a label *means*. It takes a label
and gives you the permutations — nothing more. That discipline is what keeps it a
three-line drop-in for a brand-monitoring feed, a registration watch, or a corpus
generator for testing a detector, with every consumer free to bolt the networked,
opinionated stages on top.

*Read next:* [executive-summary.md](executive-summary.md) ·
[architecture.md](architecture.md) · [userguide.md](userguide.md) ·
[../example/README.md](../example/README.md)
