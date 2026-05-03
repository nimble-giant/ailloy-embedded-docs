# Ailloy Embedded Docs

The official documentation extension for [`ailloy`](https://github.com/nimble-giant/ailloy).
Ships a rich terminal UI for browsing the project's docs with
pre-rendered [glamour](https://github.com/charmbracelet/glamour) output
across multiple widths and themes — instant page loads, no rendering
spinner, no stuttering on resize.

This binary is meant to be installed and exec'd by ailloy. The host
CLI handles download, version resolution, and lifecycle:

```bash
ailloy ext install docs        # install the extension once
ailloy docs                    # launches this TUI from now on
ailloy ext update docs         # check for a newer release
```

Standalone usage is also supported for development:

```bash
go run ./cmd/ailloy-docs                     # launch the TUI
go run ./cmd/ailloy-docs flux                # render a topic
go run ./cmd/ailloy-docs --ailloy-extension-info
```

## Architecture

```
cmd/ailloy-docs/        Extension binary entrypoint (uses extensions-sdk)
internal/tui/           Bubbletea TUI (tree nav, glamour viewport, scrollbar)
internal/styles/        Vendored from ailloy/pkg/styles for theme parity
internal/prerender/     `go generate` target: produces embedded/<style>/<width>/<slug>.md.gz
docs/                   Mirror of nimble-giant/ailloy:docs/, refreshed by the poller
embedded/               Generated artifacts; embedded via //go:embed (gitignored)
```

## Versioning

The extension has its own semver tag on its own cadence. Each release
records the ailloy version whose `docs/` it was built from, so the
ailloy host can match the right release for a given user's install.
The mapping appears in three places:

1. `ailloy-docs-version.txt` (committed source of truth)
2. Baked into the binary via `-ldflags -X main.ailloyDocsVersion=...`
3. Release notes line: `Built from ailloy docs vX.Y.Z`

## Automation

Two GitHub Actions workflows drive the repo:

- **`release.yml`** — on tag push, cross-compiles binaries for
  linux/darwin/windows × amd64/arm64, generates `checksums.txt`,
  publishes a GitHub release.
- **`poll-ailloy.yml`** — runs hourly. When a new ailloy release is
  detected, opens a PR that bumps `ailloy-docs-version.txt`, syncs
  `docs/` from that ailloy tag, and regenerates the pre-renders.
  release-please picks up the conventional-commit message and cuts
  the next extension release once the PR merges.

## License

Apache-2.0.
