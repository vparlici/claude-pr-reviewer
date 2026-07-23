package claude

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// endpoint is the Claude Messages API URL. A var (not const) so tests can point
// it at a stub server.
var endpoint = "https://api.anthropic.com/v1/messages"

const (
	version = "2023-06-01"

	// Model is the Claude model used for reviews. Exported so callers can
	// display it (e.g. in the review attribution footer).
	Model = "claude-sonnet-5"

	maxTokens = 8192
)

const systemPrompt = `You are a senior staff software engineer performing a rigorous code review on a GitHub Pull Request.

You will receive a unified diff of the changed files. Review it carefully and respond in GitHub-flavored Markdown using EXACTLY the following structure and headings:

## Summary
A short, high-level description of what this PR changes and its overall intent.

## Security & Critical Bugs
List concrete security vulnerabilities, data-loss risks, race conditions, nil/null dereferences, and logic bugs. If none are found, state that explicitly.

## Performance & Best Practices
Point out inefficiencies, unnecessary allocations, blocking calls, and deviations from idiomatic best practices for the language in question.

## Suggested Refactoring (with code snippets)
Provide actionable refactoring suggestions. Include concrete code snippets in fenced code blocks showing the improved version.

Be precise and reference file names and line context when possible. Do not invent issues; if the code is solid, say so. Keep the tone professional and constructive.`

// request is the request body for the Claude Messages API.
type request struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	System    string    `json:"system"`
	Messages  []message `json:"messages"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// response is the subset of the Claude API response we consume.
type response struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}

// Review sends the diff to the Anthropic Messages API and returns the generated
// Markdown review text.
func Review(ctx context.Context, apiKey, diff string) (string, error) {
	reqBody := request{
		Model:     Model,
		MaxTokens: maxTokens,
		System:    systemPrompt,
		Messages: []message{
			{
				Role:    "user",
				Content: "Please review the following Pull Request diff:\n\n" + diff,
			},
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", version)

	httpClient := &http.Client{Timeout: 2 * time.Minute}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	var parsed response
	if err := json.Unmarshal(body, &parsed); err != nil {
		return "", fmt.Errorf("unmarshaling response (status %d): %w", resp.StatusCode, err)
	}

	if resp.StatusCode != http.StatusOK {
		if parsed.Error != nil {
			return "", fmt.Errorf("anthropic API error (status %d): %s: %s", resp.StatusCode, parsed.Error.Type, parsed.Error.Message)
		}
		return "", fmt.Errorf("anthropic API returned status %d: %s", resp.StatusCode, string(body))
	}

	var out strings.Builder
	for _, block := range parsed.Content {
		if block.Type == "text" {
			out.WriteString(block.Text)
		}
	}

	text := strings.TrimSpace(out.String())
	if text == "" {
		return "", fmt.Errorf("anthropic API returned an empty review")
	}
	return text, nil
}
