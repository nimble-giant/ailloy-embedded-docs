// Command ailloy-docs is the docs extension binary that ailloy execs
// when the user runs `ailloy docs`. It can also run standalone for
// development and CI smoke testing.
//
// Behavior:
//
//   - With no arguments, launches the bubbletea TUI (when stdout is a TTY)
//     or prints the topics table (when piped).
//   - With a topic argument, renders that topic to stdout via glamour.
//   - Honors AILLOY_NO_COLOR / NO_COLOR environment variables.
//   - Implements the SDK protocol: --ailloy-protocol-version,
//     --ailloy-extension-info exit early with metadata.
package main

import (
	"fmt"
	"io"
	"os"

	"github.com/charmbracelet/glamour"
	clidocs "github.com/nimble-giant/ailloy-embedded-docs/docs"
	tui "github.com/nimble-giant/ailloy-embedded-docs/internal/tui"
	"github.com/nimble-giant/ailloy-extensions-sdk/pkg/extension"
	"golang.org/x/term"
)

// Populated via -ldflags at build time.
var (
	version           = "v0.0.0-dev"
	ailloyDocsVersion = "" // contents of ailloy-docs-version.txt at build time
)

const extensionName = "docs"

func main() {
	info := extension.Info{
		Name:              extensionName,
		Version:           version,
		AilloyDocsVersion: ailloyDocsVersion,
	}
	extension.HandleVersionFlags(info)

	args := os.Args[1:]

	// Handle simple flags before falling through to topic rendering.
	for _, a := range args {
		if a == "--version" || a == "-v" {
			fmt.Printf("ailloy-docs %s (built from ailloy docs %s)\n", version, ailloyDocsVersion)
			os.Exit(0)
		}
		if a == "--list" || a == "-l" {
			printTopicList(os.Stdout)
			os.Exit(0)
		}
	}

	// Topic argument → render to stdout.
	if len(args) >= 1 && args[0] != "" && args[0][0] != '-' {
		if err := renderTopic(os.Stdout, args[0]); err != nil {
			fmt.Fprintln(os.Stderr, "ailloy-docs:", err)
			os.Exit(1)
		}
		return
	}

	// No args: launch the TUI when interactive, otherwise list topics.
	if isInteractive() {
		if err := tui.Run(); err != nil {
			fmt.Fprintln(os.Stderr, "ailloy-docs:", err)
			os.Exit(1)
		}
		return
	}
	printTopicList(os.Stdout)
}

func isInteractive() bool {
	return term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd()))
}

func printTopicList(w io.Writer) {
	fmt.Fprintln(w, "Available documentation topics:")
	fmt.Fprintln(w)
	for _, t := range clidocs.List() {
		fmt.Fprintf(w, "  %-32s %s\n", t.Slug, t.Summary)
	}
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Render a topic with: ailloy docs <topic>")
}

func renderTopic(w io.Writer, slug string) error {
	body, err := clidocs.Read(slug)
	if err != nil {
		return err
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(rendererWidth()),
	)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	rendered, err := r.Render(string(body))
	if err != nil {
		return fmt.Errorf("render %s: %w", slug, err)
	}
	_, err = io.WriteString(w, rendered)
	return err
}

func rendererWidth() int {
	const def = 100
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return def
	}
	switch {
	case w > 120:
		return 120
	case w < 40:
		return 40
	default:
		return w
	}
}
