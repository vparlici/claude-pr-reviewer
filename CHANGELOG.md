# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.1.0] - 2026-07-23

### Added

- Optional `model` action input (default `claude-sonnet-5`) to choose the Claude
  model per run.
- Optional `max-tokens` action input (default `8192`) to cap the review response.
- `reviewOptions()` reads `CLAUDE_MODEL` / `CLAUDE_MAX_TOKENS` with whitespace
  trimming and a fail-soft fallback to defaults on invalid values.

### Changed

- `claude.Review` now takes an `Options` struct instead of package-level
  constants.

### Fixed

- Missing trailing newline in `action.yml` and `.github/workflows/ci.yml`.

## [1.0.0] - 2026-07-23

### Added

- Initial release: Go GitHub Action that reviews Pull Request diffs with the
  Anthropic Claude API (`claude-sonnet-5`) and posts a structured review comment.
- Structured Markdown output: Summary, Security & Critical Bugs, Performance &
  Best Practices, Suggested Refactoring.
- Smart filtering of lock files, binaries, and assets before building the diff.
- Distributed as a prebuilt distroless Docker image published to GHCR.
- CI (`go vet` / `go test` / `gofmt`) on pull requests and a release workflow
  that builds and pushes the image on version tags.

[Unreleased]: https://github.com/vparlici/claude-pr-reviewer/compare/v1.1.0...HEAD
[1.1.0]: https://github.com/vparlici/claude-pr-reviewer/compare/v1.0.0...v1.1.0
[1.0.0]: https://github.com/vparlici/claude-pr-reviewer/releases/tag/v1.0.0