# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Switcher is a Go CLI tool that launches Claude Code with different OpenAI-compatible AI providers. It manages `ANTHROPIC_BASE_URL` and `ANTHROPIC_AUTH_TOKEN` environment variables so users can switch between providers (DeepSeek, OpenRouter, Moonshot, etc.) with a single command.

Zero external dependencies — pure Go stdlib.

## Build & Development Commands

```bash
make build          # Compile binary to ./switcher
make install        # Build and install to /usr/local/bin
make clean          # Remove local binary
go build -o switcher .   # Direct build
```

No test suite exists currently.

## Architecture

The codebase is ~500 LOC across 3 internal packages:

**`main.go`** — Entry point, delegates to `cli.Run()`. Version injected via ldflags at build time.

**`internal/cli`** — Command dispatcher. `Run(args)` routes to handlers: `list`, `add`, `remove`, `edit`, `<provider-name>`. The `edit` command opens a temp JSON file in `$EDITOR`. Platform-specific editor defaults via build tags (`editor_unix.go` / `editor_windows.go`).

**`internal/config`** — Reads/writes `~/.switcher.json`. The `Provider` struct holds `BaseURL`, `APIKey`, `Model`, `Description`, and `Env` (extra env vars). `Load()` creates default providers on first run.

**`internal/runner`** — Finds `claude` in PATH, sets env vars, launches Claude. On Unix uses `syscall.Exec()` (process replacement). On Windows uses `exec.Command()` as subprocess and also checks for `claude.cmd` (npm global installs). Platform split via `exec_unix.go` / `exec_windows.go`.

## Key Design Decisions

- **Platform abstraction via build tags**: `//go:build windows` vs `//go:build !windows` for process execution and editor selection
- **Process replacement on Unix**: `syscall.Exec()` replaces switcher's process with Claude (zero overhead, clean terminal)
- **Config as flat JSON**: `~/.switcher.json` maps provider names to Provider structs
- **Env var management**: `setEnv`/`removeEnv` helpers manipulate `os.Environ()` list directly before exec

## Go Style

Follow Effective Go conventions: `gofmt`, MixedCaps naming, explicit error handling, small interfaces. No underscores in names.

## Releasing

Releases use GoReleaser triggered by pushing annotated git tags (`git tag -a vX.Y.Z`). The GitHub Actions workflow builds multi-platform binaries and signs artifacts with Cosign.
