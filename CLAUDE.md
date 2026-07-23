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

## AI Integration & Prompting Guidelines
- **Model Target:** Default to `claude-sonnet-5` (or the latest stable Sonnet release). The model and token budget are configurable per-run via the `model` / `max-tokens` action inputs; defaults live in `internal/claude` (`DefaultModel`, `DefaultMaxTokens`).
- **Output Expectations:** Prompts must enforce structured Markdown output covering:
    1. Summary
    2. Security & Critical Bugs
    3. Performance & Best Practices
    4. Suggested Refactoring (with code snippets)
- **Token Efficiency:** Always filter out irrelevant files (e.g., `.lock`, `.json`, `.png`, `.svg`) before building the diff string to minimize API latency and token consumption.