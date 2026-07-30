// Nested module: the twister ↔ snare differential harness. It lives in its own
// module so the root twister module stays zero-dependency — this harness is the
// only thing that imports snare. Run it with both repos checked out as siblings
// (the replace directives point at ../ and ../../snare).
module github.com/netstar-labs/twister/diff

go 1.25.0

require (
	github.com/netstar-labs/snare v0.0.0
	github.com/netstar-labs/twister v0.0.0
)

replace (
	github.com/netstar-labs/snare => ../../snare
	github.com/netstar-labs/twister => ../
)
