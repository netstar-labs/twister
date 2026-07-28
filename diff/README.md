# diff — twister ↔ twist differential harness

A **nested module** that cross-validates the two libraries: for every seed, each
edit-based fuzzer variant produced by `twister` must be detected by `twist` as a
near-miss at **distance 1**. Because each edit-based fuzzer makes exactly one
Damerau-Levenshtein edit, this proves both libraries at once — `twister` generates
the candidate, `twist` confirms it is one edit away.

It lives in its own module so the **root `twister` module stays
zero-dependency** — this harness is the only thing that imports `twist`. The
`replace` directives in [go.mod](go.mod) point at the sibling checkouts (`../` and
`../../twist`), so run it with both repos checked out side by side:

```sh
cd diff
GOWORK=off go test ./...
```

Root CI does not run this module (it is outside `./...` of the root module); it is a
local cross-repo harness, run on demand.
