package twister

import "testing"

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
