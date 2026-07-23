# CLAUDE.md - Development & Architecture Guide

## Overview
This repository hosts a lightweight, high-performance GitHub Action built in **Go 1.22+**. It automatically performs automated AI code reviews on Pull Requests using the Anthropic Claude API.

## Core Commands
- **Local Build:** `go build -o reviewer .`
- **Run Tests:** `go test ./...`
- **Run Linter / Static Analysis:** `go vet ./...` (or `golangci-lint run`)
- **Format Code:** `gofmt -w .`
- **Tidy Dependencies:** `go mod tidy`

## Architecture & Coding Standards
- **Idiomatic Go:** Follow standard Go conventions (*Effective Go*). Keep functions focused, readable, and simple.
- **Explicit Error Handling:** Check every error explicitly (`if err != nil`). Never use `panic()` in production logic; exit cleanly via `os.Exit(1)` inside `main()` if fatal dependencies are missing.
- **Minimal Dependencies:** Keep external dependencies to an absolute minimum for zero-overhead CI/CD execution.
    - **GitHub API:** Use the official `github.com/google/go-github/v60` SDK.
    - **Anthropic API:** Use native HTTP requests via the standard `net/http` package (avoid unnecessary third-party wrappers).
- **Environment Context:** The application relies on standard GitHub Action environment variables: `GITHUB_TOKEN`, `ANTHROPIC_API_KEY`, and `GITHUB_EVENT_PATH`. Optional overrides `CLAUDE_MODEL` and `CLAUDE_MAX_TOKENS` (mapped from the `model` and `max-tokens` action inputs) fall back to the defaults in `internal/claude` when unset or invalid.

## Packaging & Distribution
- **Docker container action (not composite `go run`):** The action is packaged as a Docker container action referencing a prebuilt image (`ghcr.io/vparlici/claude-pr-reviewer`), not a composite action that compiles at run time.
    - **Why not composite `go run`/`go build`:** it would recompile the binary and re-download modules on *every* PR (~30-60s of wasted CI time per run, plus a `setup-go` step).
    - **Why not a committed binary:** shipping per-platform binaries in the repo is ugly and hard to maintain.
    - **Why Docker wins:** compile once at release time, ship a tiny static binary in a `distroless/static` image, and at run time only pull + exec — zero compilation, zero dependency install. Reproducible (pinned base image, bundled CA certs) and keeps the repo free of build artifacts.
    - **Trade-off:** Docker actions run on Linux runners only. Acceptable here — the action targets `ubuntu-latest`.
- **Build flags:** `CGO_ENABLED=0` for a fully static binary; `-trimpath -ldflags="-s -w"` to strip and shrink. Build the whole module (`go build .`), not `main.go`, so the `internal/` packages are included.
- **Release flow:** pushing a `v*` tag triggers `.github/workflows/release.yml` (test → build → push image, tagged `vX.Y.Z` / `vX.Y` / `vX` / `latest`). The moving major **git** tag `v1` must be repointed manually (`git tag -f v1 <release> && git push -f origin v1`) so `uses: ...@v1` picks up the new `action.yml`.

## AI Integration & Prompting Guidelines
- **Model Target:** Default to `claude-sonnet-5` (or the latest stable Sonnet release). The model and token budget are configurable per-run via the `model` / `max-tokens` action inputs; defaults live in `internal/claude` (`DefaultModel`, `DefaultMaxTokens`).
- **Output Expectations:** Prompts must enforce structured Markdown output covering:
    1. Summary
    2. Security & Critical Bugs
    3. Performance & Best Practices
    4. Suggested Refactoring (with code snippets)
- **Token Efficiency:** Always filter out irrelevant files (e.g., `.lock`, `.json`, `.png`, `.svg`) before building the diff string to minimize API latency and token consumption.