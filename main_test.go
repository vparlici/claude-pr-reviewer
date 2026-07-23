package main

import (
	"testing"

	"github.com/vparlici/claude-pr-reviewer/internal/claude"
)

func TestReviewOptions(t *testing.T) {
	tests := []struct {
		name       string
		model      string
		maxTokens  string
		wantModel  string
		wantTokens int
	}{
		{"defaults", "", "", claude.DefaultModel, claude.DefaultMaxTokens},
		{"custom model", "claude-haiku-4-5", "", "claude-haiku-4-5", claude.DefaultMaxTokens},
		{"custom tokens", "", "4096", claude.DefaultModel, 4096},
		{"whitespace trimmed", "  claude-opus-4-8  ", " 2048 ", "claude-opus-4-8", 2048},
		{"invalid tokens falls back", "", "not-a-number", claude.DefaultModel, claude.DefaultMaxTokens},
		{"negative tokens falls back", "", "-1", claude.DefaultModel, claude.DefaultMaxTokens},
		{"zero tokens falls back", "", "0", claude.DefaultModel, claude.DefaultMaxTokens},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CLAUDE_MODEL", tt.model)
			t.Setenv("CLAUDE_MAX_TOKENS", tt.maxTokens)

			got := reviewOptions()
			if got.Model != tt.wantModel {
				t.Errorf("Model = %q, want %q", got.Model, tt.wantModel)
			}
			if got.MaxTokens != tt.wantTokens {
				t.Errorf("MaxTokens = %d, want %d", got.MaxTokens, tt.wantTokens)
			}
		})
	}
}
