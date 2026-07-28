package twister

import "testing"

// BenchmarkPermute measures the full core set at representative label lengths.
func BenchmarkPermute(b *testing.B) {
	for _, label := range []string{"go", "google", "verylongbrandname"} {
		b.Run(label, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = Permute(label)
			}
		})
	}
}

// BenchmarkPermuteWithTLD includes the tld-swap fuzzer (needs a TLD list).
func BenchmarkPermuteWithTLD(b *testing.B) {
	o := Options{TLDs: []string{"com", "net", "org", "io", "co"}}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = PermuteWith("google", o)
	}
}

// BenchmarkFuzzers records a per-fuzzer baseline on a fixed label, so a
// regression in any single technique is visible in isolation.
func BenchmarkFuzzers(b *testing.B) {
	const label = "cloudflare"
	tlds := []string{"com", "net", "org"}
	for _, name := range FuzzerNames {
		o := Options{Fuzzers: []string{name}, TLDs: tlds}
		b.Run(name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				_ = PermuteWith(label, o)
			}
		})
	}
}
