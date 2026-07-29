// Command twister generates typosquat permutations of a domain label.
//
//	twister permute [-f a,b,c] [-tld com,net,org] [-words login,secure] [-maxedits 2] [labels...]
//	twister version
//
// -f restricts to a comma-separated set of fuzzers (default all; see the twister
// package's FuzzerNames). -tld feeds the tld-swap fuzzer, -words feeds combosquat,
// and -maxedits (>=2) enables the aggressive fuzzer. Labels are the command
// arguments, or one per line on stdin if none are given. Each variant prints as:
// label <tab> variant <tab> fuzzer <tab> editcount.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/netstar-labs/twister"
)

// stamped by build/twister via -ldflags -X.
var (
	version = "dev"
	build   = "none"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	var err error
	switch os.Args[1] {
	case "permute":
		err = permute(os.Args[2:])
	case "version", "-version", "--version", "-v":
		fmt.Printf("twister %s (%s)\n", version, build)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "twister:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: twister <permute|version> [flags] [labels...]")
	os.Exit(2)
}

func permute(args []string) error {
	fs := flag.NewFlagSet("permute", flag.ExitOnError)
	fuzzers := fs.String("f", "", "comma-separated fuzzers (default all)")
	tlds := fs.String("tld", "", "comma-separated TLDs for the tld-swap fuzzer")
	words := fs.String("words", "", "comma-separated keywords for the combosquat fuzzer")
	maxEdits := fs.Int("maxedits", 0, "max positions the aggressive fuzzer substitutes at once (>=2)")
	fs.Parse(args)

	o := twister.Options{
		Fuzzers:  splitList(*fuzzers),
		TLDs:     splitList(*tlds),
		Words:    splitList(*words),
		MaxEdits: *maxEdits,
	}

	w := bufio.NewWriter(os.Stdout)
	defer w.Flush()
	emit := func(label string) {
		if label == "" {
			return
		}
		for _, v := range twister.PermuteWith(label, o) {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", label, v.Name, v.Fuzzer, v.EditCount)
		}
	}

	if labels := fs.Args(); len(labels) > 0 {
		for _, l := range labels {
			emit(l)
		}
		return nil
	}
	// No arguments: read one label per line from stdin.
	sc := newLineScanner(os.Stdin)
	for sc.Scan() {
		emit(strings.TrimSpace(sc.Text()))
	}
	return sc.Err()
}

// splitList parses a comma-separated flag into a trimmed, non-empty slice (nil if
// the flag is empty).
func splitList(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	var out []string
	for _, p := range strings.Split(s, ",") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

// newLineScanner returns a bufio.Scanner over r that tolerates lines up to 1 MiB.
func newLineScanner(r io.Reader) *bufio.Scanner {
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	return sc
}
