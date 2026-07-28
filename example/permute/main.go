// Runnable example: permute a brand label and group the variants by fuzzer.
//
//	go run ./example/permute
package main

import (
	"fmt"
	"sort"

	"github.com/netstar-labs/twister"
)

func main() {
	const brand = "paypal"

	// Full core set — every fuzzer except tld-swap (no TLD list supplied).
	all := twister.Permute(brand)
	fmt.Printf("Permute(%q): %d variants\n\n", brand, len(all))

	byFuzzer := map[string][]string{}
	for _, v := range all {
		byFuzzer[v.Fuzzer] = append(byFuzzer[v.Fuzzer], v.Name)
	}
	for _, name := range twister.FuzzerNames {
		vs := byFuzzer[name]
		if len(vs) == 0 {
			continue
		}
		sort.Strings(vs)
		show := vs
		if len(show) > 6 {
			show = show[:6]
		}
		fmt.Printf("  %-14s %2d  %v\n", name, len(vs), show)
	}

	// tld-swap needs a caller-supplied TLD list (twister ships none).
	fmt.Println("\nPermuteWith tld-swap {com,net,org,co}:")
	for _, v := range twister.PermuteWith(brand, twister.Options{
		Fuzzers: []string{"tld-swap"},
		TLDs:    []string{"com", "net", "org", "co"},
	}) {
		fmt.Printf("  %s\n", v.Name)
	}
}
