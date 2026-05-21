# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

All build/test flows go through the Makefile, which pins `GOCACHE` to `.go-build-cache/` and sets the CGO flags Fyne needs. Prefer it over raw `go` commands:

- `make build` — host binary into `bin/lux-app` (CGO required).
- `make test` — runs `go test ./...` with the right cache and CGO flags. Single test: `GOCACHE="$(pwd)/.go-build-cache" go test ./internal/luxdownloader -run TestName`.
- `go run .` — fastest dev loop (skips packaging).
- `make package-native` — Fyne-packaged artifact (`.app` / `.exe` / `.tar.*`) for the host platform into `dist/`.
- `make dist` — cross-compiles all four desktop targets (`darwin-amd64`, `darwin-arm64`, `windows-amd64`, `windows-arm64`) using `zig cc`/`zig c++`. Requires `zig` and, for macOS targets, the host Xcode SDK (`xcrun --sdk macosx --show-sdk-path`).
- `make clean` — wipes `bin/`, `dist/`, all Go/Zig caches, and stray packaging artifacts.

Linux desktop packages need X11/OpenGL dev headers and must be built on Linux (`make package-native`); `make dist-linux` just prints that reminder.

Windows builds funnel through `scripts/zig-wrap`, which strips flags Fyne's cgo emits but Zig's clang rejects (`-Wp,-D_FORTIFY_SOURCE=*`, `-fstack-protector-strong`). If a new Fyne/Go release adds another rejected flag, add it to the case there rather than working around it elsewhere.

## Architecture

Three layers, top-down:

1. **Fyne UI (`main.go` + `config.go`, `package main`)** — builds the window, wires widgets to a `downloadConfig`, persists every field via `fyne.Preferences` (keys defined in `config.go`), and runs the download in a goroutine. All UI mutations from background goroutines must go through `fyne.Do(...)` — see `appendLog` and the `updateProgress` callback in `main.go` for the pattern. `version` is set via `-ldflags -X main.version=...` from the Makefile.

2. **`internal/luxrunner`** — thin orchestrator. `Run(Config)` normalizes inputs, calls `request.SetOptions` (global lux state — be aware concurrent `Run`s would race on it), iterates URLs through `extractors.Extract`, and hands each result to a `luxdownloader.Downloader`. `register.go` is the side-effect import list for every lux extractor the app supports; adding a new site means adding an import there. `luxrunner` re-exports the `ProgressEvent` / `ProgressPhase` types from `luxdownloader` so `main.go` doesn't depend on the lower package directly.

3. **`internal/luxdownloader`** — a vendored fork of lux's `downloader` package, modified to emit progress instead of printing a progress bar. The key addition is `progress.go`'s `progressTracker`: byte-level writes flow through `progressWriter` → `tracker.add()` → throttled `callback(event)` (default 100ms, configured via `ProgressThrottle`). Phase transitions (`ProgressExtracting` → `Downloading` → `Merging`/`Skipped` → `Finished`) are pushed via `setPhase` / `complete`. When touching `downloader.go`, keep the multi-thread resume logic (`.part` files with `FilePartMeta` headers, `.download` temp suffix) intact — interrupting and restarting a download relies on it.

The Fyne app ID (`github.com.tbxark.lux-app`) is duplicated in `main.go`, `Makefile` (`APP_ID`), and `FyneApp.toml`; keep them in sync. `FyneApp.toml` and `Icon.png` are consumed by `fyne package` during `make package`.

## Cookie handling quirk

`luxrunner.cookieValue` treats the cookie field as a **file path if it stat()s as a regular file**, otherwise as a literal cookie header. Tests and UI hints should preserve that dual behavior.
