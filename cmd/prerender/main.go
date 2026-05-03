// Command prerender runs the build-time pre-render generator. Invoke
// via `go generate ./...` from the repo root, or directly:
//
//	go run ./cmd/prerender -out internal/embedded -v
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/nimble-giant/ailloy-embedded-docs/internal/prerender"
)

func main() {
	var (
		out     = flag.String("out", "internal/embedded", "output directory for generated artifacts")
		styles  = flag.String("styles", "", "comma-separated glamour style names (default: dark,light)")
		widths  = flag.String("widths", "", "comma-separated word-wrap widths (default: 60,80,100,120,140)")
		verbose = flag.Bool("v", false, "print one line per generated file")
	)
	flag.Parse()

	opts := prerender.Options{
		OutputDir: *out,
		Verbose:   *verbose,
	}
	if *styles != "" {
		for _, s := range strings.Split(*styles, ",") {
			s = strings.TrimSpace(s)
			if s != "" {
				opts.Styles = append(opts.Styles, s)
			}
		}
	}
	if *widths != "" {
		for _, w := range strings.Split(*widths, ",") {
			w = strings.TrimSpace(w)
			if w == "" {
				continue
			}
			n, err := strconv.Atoi(w)
			if err != nil {
				fmt.Fprintf(os.Stderr, "prerender: invalid width %q\n", w)
				os.Exit(2)
			}
			opts.Widths = append(opts.Widths, n)
		}
	}

	if err := prerender.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, "prerender:", err)
		os.Exit(1)
	}
}
