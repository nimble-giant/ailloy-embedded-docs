package tui

import (
	"github.com/muesli/termenv"
	"github.com/nimble-giant/ailloy-embedded-docs/internal/embedded"
)

// loadPrerendered tries to pull a pre-rendered glamour output for slug
// at the closest available width and the user's active theme. Returns
// the rendered content and ok=true on a hit; ok=false signals a fall-
// back to live rendering.
func loadPrerendered(slug string, requestedWidth int, dark bool) (string, bool) {
	width := embedded.SnapWidth(requestedWidth)
	if width == 0 {
		return "", false
	}
	style := "dark"
	if !dark {
		style = "light"
	}
	body, err := embedded.Load(slug, style, width)
	if err != nil || body == nil {
		return "", false
	}
	return string(body), true
}

// isDarkBackground reports whether the current terminal has a dark
// background — used to pick which pre-rendered style to load. Defaults
// to dark when the answer can't be determined; that's the safer
// majority case for terminal users.
func isDarkBackground() bool {
	return termenv.HasDarkBackground()
}
