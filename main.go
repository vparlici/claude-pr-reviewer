// Command claude-pr-reviewer is a GitHub Action that reviews Pull Request diffs
// using the Anthropic Claude API and posts a structured review comment.
package main

import (
	"context"
	"fmt"
	"os"
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

	review, err := claude.Review(ctx, anthropicKey, diff)
	if err != nil {
		return fmt.Errorf("requesting review from Claude: %w", err)
	}

	if err := github.PostComment(ctx, gh, pr, review, claude.Model); err != nil {
		return fmt.Errorf("posting review comment: %w", err)
	}

	fmt.Println("Review posted successfully.")
	return nil
}
