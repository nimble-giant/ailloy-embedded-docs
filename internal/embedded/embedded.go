// Package embedded exposes the build-time pre-rendered glamour
// outputs to the TUI. Files are produced by `cmd/prerender` and
// stored under `<style>/<width>/<slug>.md.gz` (with "/" in slugs
// replaced by "__" so subdirectory names survive the flat
// filesystem layout).
//
// At runtime the TUI:
//
//   - picks a style based on the user's terminal background
//     (auto-detected via termenv, or forced via NO_COLOR) and
//   - snaps the window's word-wrap width to the nearest pre-rendered
//     width.
//
// Lookups that miss fall through to a live glamour render in the TUI.
package embedded

import (
	"bytes"
	"compress/gzip"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
)

// embeddedFS includes everything under internal/embedded/. Empty
// initial trees cause `//go:embed` to fail at build time, so we keep a
// MARKER file alongside the generated artifacts.
//
//go:embed all:dark all:light MARKER
var embeddedFS embed.FS

// FS exposes the raw embedded filesystem for tests / introspection.
func FS() fs.FS { return embeddedFS }

// AvailableWidths returns the widths the binary was built with, sorted
// ascending. Calculated by walking the dark/ tree at startup.
func AvailableWidths() []int {
	entries, err := fs.ReadDir(embeddedFS, "dark")
	if err != nil {
		return nil
	}
	out := make([]int, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		w, err := strconv.Atoi(e.Name())
		if err != nil {
			continue
		}
		out = append(out, w)
	}
	sort.Ints(out)
	return out
}

// SnapWidth picks the available width closest to the requested width.
// Returns 0 if no widths are embedded.
func SnapWidth(want int) int {
	widths := AvailableWidths()
	if len(widths) == 0 {
		return 0
	}
	best := widths[0]
	bestDelta := abs(widths[0] - want)
	for _, w := range widths[1:] {
		d := abs(w - want)
		if d < bestDelta {
			best = w
			bestDelta = d
		}
	}
	return best
}

// Load returns the rendered output for slug at the given style + width.
// Returns nil + nil if the file isn't embedded; callers can fall back
// to live rendering. Returns an error only on unexpected I/O failure.
func Load(slug, style string, width int) ([]byte, error) {
	if style != "dark" && style != "light" {
		return nil, fmt.Errorf("unknown style %q (want dark|light)", style)
	}
	safe := strings.ReplaceAll(slug, "/", "__") + ".md.gz"
	p := path.Join(style, strconv.Itoa(width), safe)
	data, err := embeddedFS.ReadFile(p)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("gunzip %s: %w", p, err)
	}
	defer func() { _ = gz.Close() }()
	out, err := io.ReadAll(gz)
	if err != nil {
		return nil, fmt.Errorf("read %s: %w", p, err)
	}
	return out, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
