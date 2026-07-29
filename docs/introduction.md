# Meet twister — the permutation, not the near-miss

Every phishing kit begins with a lie told in a domain name. `paypa1.com` for
`paypal.com`; `gooogle`; `pаypal` with a look-alike Cyrillic `а` you would swear
was the real letter. Before a defender can catch one of those, someone has to
enumerate the space of what a squatter *would* register. That is the offensive,
generative half of the problem — give it a brand and it *generates* the thousands of
plausible permutations — and it is exactly the half `twist` does not do. twister is
the counter to `twist`: hand it a label and it answers one question: *what could squat
this?*

## What it actually is

twister is a pure-Go, zero-dependency **permutation generator**. You give it a
registrable label — `paypal`, `google`, whatever you are protecting — and it returns
the set of look-alike labels an attacker might register, each tagged with the
technique that produced it and how many edits it lies from the seed. It walks
twelve single-edit fuzzers: dropping a character
(`paypal`→`papal`), doubling one, swapping neighbours (`papyal`), a keyboard slip to
an adjacent key, a hyphen or a dot slipped in, a vowel swapped, a Unicode homoglyph,
a letter typed as a look-alike digit (`paypal`→`p4ypal`), a single flipped bit — plus,
just off that core, a swap to a different TLD. Each edit-based fuzzer makes *exactly
one* change, so every variant is one edit-distance step from the seed — which is what
makes twister the mirror image of `twist`: what twister generates, twist detects.

On top of that core sits an opt-in v0.2 layer of four **multi-edit** fuzzers —
combosquat (`paypal-login`), aggressive multi-substitution (`p4yp4l`),
multi-character homoglyphs (`rn`→`m`), and homophones (`ph`↔`f`). All are off by
default — combosquat and aggressive are data-gated on an `Options` field (like
tld-swap), multi-homoglyph and homophone you enable by name — and every variant is
tagged with its edit count so it routes to the right detector, never diluting the
tight dist-1 core that `Permute` returns.

## The line it will not cross

A full squat-hunting pipeline does more than generate: it then resolves every
candidate, runs whois, geoip, banner grabs, MX and port checks. twister does **none**
of that. It is a pure
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
