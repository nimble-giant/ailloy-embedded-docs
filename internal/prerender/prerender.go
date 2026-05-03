// Package prerender is the build-time generator that walks every
// embedded markdown topic, renders it through glamour at each
// (style, width) combination, and writes gzipped output into
// `internal/embedded/<style>/<width>/<slug>.md.gz`.
//
// The TUI loads these at runtime so reading-pane updates are
// effectively instantaneous. Run via `go generate ./...` from the
// repository root, or invoke the cmd/prerender binary directly.
package prerender

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	clidocs "github.com/nimble-giant/ailloy-embedded-docs/docs"
)

// DefaultStyles is the set of glamour style names this generator
// produces output for. Pre-rendering both styles lets the TUI pick at
// runtime based on the user's terminal background without paying for a
// glamour render.
var DefaultStyles = []string{"dark", "light"}

// DefaultWidths is the set of word-wrap widths this generator produces
// output for. The TUI snaps to the nearest at runtime.
var DefaultWidths = []int{60, 80, 100, 120, 140}

// Options configures a generation run.
type Options struct {
	// OutputDir is the root under which `<style>/<width>/<slug>.md.gz`
	// files are written. Created if missing.
	OutputDir string
	// Styles overrides DefaultStyles when non-empty.
	Styles []string
	// Widths overrides DefaultWidths when non-empty.
	Widths []int
	// Verbose, when true, prints one line per generated file.
	Verbose bool
	// Out receives progress messages when Verbose is true. Defaults to
	// os.Stdout.
	Out io.Writer
}

// Run executes a full pre-render pass.
func Run(opts Options) error {
	if opts.OutputDir == "" {
		return fmt.Errorf("OutputDir is required")
	}
	if len(opts.Styles) == 0 {
		opts.Styles = DefaultStyles
	}
	if len(opts.Widths) == 0 {
		opts.Widths = DefaultWidths
	}
	if opts.Out == nil {
		opts.Out = os.Stdout
	}

	if err := os.MkdirAll(opts.OutputDir, 0o750); err != nil {
		return fmt.Errorf("create output dir: %w", err)
	}

	topics := clidocs.List()
	if len(topics) == 0 {
		return fmt.Errorf("no embedded topics to render")
	}

	for _, style := range opts.Styles {
		for _, width := range opts.Widths {
			renderer, err := glamour.NewTermRenderer(
				glamour.WithStandardStyle(style),
				glamour.WithWordWrap(width),
			)
			if err != nil {
				return fmt.Errorf("create renderer (style=%s, width=%d): %w", style, width, err)
			}

			styleDir := filepath.Join(opts.OutputDir, style, fmt.Sprintf("%d", width))
			if err := os.MkdirAll(styleDir, 0o750); err != nil {
				_ = renderer.Close()
				return fmt.Errorf("create style/width dir: %w", err)
			}

			for _, t := range topics {
				body, err := clidocs.Read(t.Slug)
				if err != nil {
					_ = renderer.Close()
					return fmt.Errorf("read %s: %w", t.Slug, err)
				}
				rendered, err := renderer.Render(string(body))
				if err != nil {
					_ = renderer.Close()
					return fmt.Errorf("render %s (style=%s, width=%d): %w",
						t.Slug, style, width, err)
				}

				safe := strings.ReplaceAll(t.Slug, "/", "__")
				dest := filepath.Join(styleDir, safe+".md.gz")
				if err := writeGzip(dest, rendered); err != nil {
					_ = renderer.Close()
					return err
				}
				if opts.Verbose {
					fmt.Fprintf(opts.Out, "  %s/%d  %s\n", style, width, t.Slug)
				}
			}
			_ = renderer.Close()
		}
	}
	return nil
}

func writeGzip(path, content string) error {
	var buf bytes.Buffer
	gz, err := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	if err != nil {
		return err
	}
	if _, err := io.WriteString(gz, content); err != nil {
		_ = gz.Close()
		return err
	}
	if err := gz.Close(); err != nil {
		return err
	}
	return os.WriteFile(path, buf.Bytes(), 0o644) //nolint:gosec // generated artifacts
}
