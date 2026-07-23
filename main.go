// Command claude-pr-reviewer is a GitHub Action that reviews Pull Request diffs
// using the Anthropic Claude API and posts a structured review comment.
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vparlici/claude-pr-reviewer/internal/claude"
	"github.com/vparlici/claude-pr-reviewer/internal/github"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "claude-pr-reviewer: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is required")
	}
	anthropicKey := os.Getenv("ANTHROPIC_API_KEY")
	if anthropicKey == "" {
		return fmt.Errorf("ANTHROPIC_API_KEY environment variable is required")
	}
	eventPath := os.Getenv("GITHUB_EVENT_PATH")
	if eventPath == "" {
		return fmt.Errorf("GITHUB_EVENT_PATH environment variable is required")
	}

	pr, err := github.ParseEvent(eventPath)
	if err != nil {
		return fmt.Errorf("parsing event payload: %w", err)
	}
	fmt.Printf("Reviewing %s/%s PR #%d\n", pr.Owner, pr.Repo, pr.Number)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	gh := github.NewClient(githubToken)

	diff, err := github.BuildDiff(ctx, gh, pr)
	if err != nil {
		return fmt.Errorf("building diff: %w", err)
	}
	if strings.TrimSpace(diff) == "" {
		fmt.Println("No reviewable changes found after filtering; skipping.")
		return nil
	}

	opts := reviewOptions()
	review, err := claude.Review(ctx, anthropicKey, diff, opts)
	if err != nil {
		return fmt.Errorf("requesting review from Claude: %w", err)
	}

	if err := github.PostComment(ctx, gh, pr, review, opts.Model); err != nil {
		return fmt.Errorf("posting review comment: %w", err)
	}

	fmt.Println("Review posted successfully.")
	return nil
}

// reviewOptions builds the review configuration from the CLAUDE_MODEL and
// CLAUDE_MAX_TOKENS environment variables, falling back to defaults (and
// warning) when they are unset or invalid.
func reviewOptions() claude.Options {
	opts := claude.Options{
		Model:     claude.DefaultModel,
		MaxTokens: claude.DefaultMaxTokens,
	}

	if v := os.Getenv("CLAUDE_MODEL"); v != "" {
		opts.Model = v
	}

	if v := os.Getenv("CLAUDE_MAX_TOKENS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			opts.MaxTokens = n
		} else {
			fmt.Fprintf(os.Stderr, "warning: invalid CLAUDE_MAX_TOKENS %q; using default %d\n", v, claude.DefaultMaxTokens)
		}
	}

	return opts
}
