# 🤖 Claude PR Reviewer

[![GitHub Marketplace](https://img.shields.io/badge/Marketplace-Claude%20PR%20Reviewer-blueviolet?logo=github)](https://github.com/marketplace)
[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](./LICENSE)
[![Powered by Claude](https://img.shields.io/badge/Powered%20by-Claude%20Sonnet%205-D97757)](https://www.anthropic.com/claude)

An ultra-fast, zero-dependency GitHub Action written in **Go** that automatically reviews your Pull Requests with the **Anthropic Claude API** and posts a structured, actionable review comment.

---

## Features

- **Fast & lightweight** — a single Go binary, no bloated runtimes.
- **Powered by Claude Sonnet 5** — deep, context-aware code analysis.
- **Structured output** — every review follows a consistent Markdown format:
  - Summary
  - Security & Critical Bugs
  - Performance & Best Practices
  - Suggested Refactoring (with code snippets)
- **Smart filtering** — automatically skips lock files, binaries, and assets to save tokens.
- **Secure** — your API key lives in encrypted repository secrets.

---

## Why claude-pr-reviewer?

Most PR bots fall into one of two camps: **noisy AI SaaS tools** that charge per seat and route your code through their servers, or **static linters** that only understand syntax. This action sits in the gap — deep AI reasoning, no middleman, no spam.

### Minimal CI overhead

No `node_modules`, no `pip install`, no SaaS proxy. The action ships as a small **statically-linked Go binary** in a distroless image — nothing to install at run time, and execution is sub-second. The only real wait is Claude's API response, not your dependency tree.

### Direct API pricing — no per-seat tax

100% open source. You bring your own Anthropic key and pay **per token, directly to Anthropic** — typically pennies per review. No subscriptions, no per-developer seat limits that punish you for growing the team.

### Your code never touches a third-party server

The action runs **inside your own GitHub Actions runner**. The diff goes straight from your runner to Anthropic's official API over TLS — no intermediate logging servers, no third-party storage. That's a much shorter compliance conversation than "we route our private codebase through vendor X."

### Logic, not just syntax

Linters (`eslint`, `golangci-lint`) check style and syntax. They can't reason about business logic, edge cases, cross-file context, or subtle concurrency bugs. **Claude Sonnet 5** does — near-Opus quality on code, at Sonnet cost.

### High signal, zero fluff

The system prompt explicitly forbids filler. No "Great code!", no whitespace nitpicks. Every comment is an actionable finding — a bug, a security risk, a performance issue, or a concrete refactor. If the PR is clean, it says so briefly and gets out of the way.

### Comparison

| | **claude-pr-reviewer** | Commercial AI SaaS bots | Traditional linters |
| --- | --- | --- | --- |
| **Pricing** | Free / pay-per-token | $15–$30 / user / mo | Free |
| **CI overhead** | Tiny prebuilt image, no install step | Heavy runtime or SaaS proxy | Fast |
| **Logic & security context** | High (Claude Sonnet 5) | High | Low (syntax only) |
| **Data privacy** | Direct to Anthropic | Through 3rd-party servers | Local |
| **Setup** | A few lines of YAML | OAuth / app integration | Manual config |

---

## Setup

### 1. Add your Anthropic API key as a repository secret

1. Go to your repository → **Settings** → **Secrets and variables** → **Actions**.
2. Click **New repository secret**.
3. Name it `ANTHROPIC_API_KEY` and paste your key from the [Anthropic Console](https://console.anthropic.com/).

### 2. Create the workflow file

Add the following to `.github/workflows/review.yml`:

```yaml
name: Claude PR Review

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read
  pull-requests: write

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - name: Claude PR Reviewer
        uses: vparlici/claude-pr-reviewer@v1
        with:
          anthropic-api-key: ${{ secrets.ANTHROPIC_API_KEY }}
          github-token: ${{ github.token }}
          # Optional overrides:
          # model: claude-haiku-4-5
          # max-tokens: "4096"
```

That's it. Open a Pull Request and Claude will post its review automatically.

---

## Inputs

| Input               | Required | Default              | Description                                              |
| ------------------- | -------- | -------------------- | -------------------------------------------------------- |
| `anthropic-api-key` | Yes      | —                    | Your Anthropic API key (store it as a secret).           |
| `github-token`      | No       | `${{ github.token }}` | Token used to read PR files and post the review comment. |
| `model`             | No       | `claude-sonnet-5`    | Claude model ID for the review (e.g. `claude-haiku-4-5` for lower cost, `claude-opus-4-8` for max depth). |
| `max-tokens`        | No       | `8192`               | Max tokens for the review response (thinking + output). Invalid values fall back to the default. |

> **Note:** The workflow needs `pull-requests: write` permission so the action can post the review comment.

---

## How It Works

1. Reads the GitHub event payload to determine the repository and PR number.
2. Fetches the changed files via the official `go-github` SDK.
3. Filters out non-reviewable files (`.lock`, `.png`, `.svg`, `package-lock.json`, …).
4. Sends the filtered diff to the Claude Messages API with a strict review system prompt.
5. Posts the generated review back to the Pull Request as a comment.

---

## Runtime & Publishing

This action runs as a **Docker container action** using a pre-built image, so there is **no compilation at run time** — the runner simply pulls the image and executes it.

The image is built and pushed to **GitHub Container Registry (GHCR)** automatically by [`.github/workflows/release.yml`](./.github/workflows/release.yml) whenever you push a version tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

This publishes `ghcr.io/vparlici/claude-pr-reviewer` with the tags `v1.0.0`, `v1.0`, `v1`, and `latest`. The `action.yml` references the `:v1` tag, so consumers pinning `@v1` always get the latest compatible build.

> **First release:** make the GHCR package public (repo → **Packages** → package → **Package settings** → **Change visibility**) so other repositories can pull it.

## Local Development

```bash
go build -o reviewer .                # build
go test ./...                         # test
go vet ./...                          # static analysis
gofmt -w .                            # format
docker build -t claude-pr-reviewer .  # build the container image locally
```

See [`CLAUDE.md`](./CLAUDE.md) for full architecture and contribution guidelines.

---

## License

Released under the [MIT License](./LICENSE).