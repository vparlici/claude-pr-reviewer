package github

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	gogithub "github.com/google/go-github/v60/github"
)

// newTestClient returns a go-github client whose requests are routed to a test
// server running handler, plus a cleanup function.
func newTestClient(t *testing.T, handler http.Handler) *gogithub.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)

	client := gogithub.NewClient(nil)
	base, err := url.Parse(srv.URL + "/")
	if err != nil {
		t.Fatalf("parsing test server URL: %v", err)
	}
	client.BaseURL = base
	return client
}

func TestShouldSkip(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"main.go", false},
		{"src/app/handler.ts", false},
		{"README.md", false},
		{"go.sum", true},
		{"package-lock.json", true},
		{"frontend/pnpm-lock.yaml", true},
		{"assets/logo.png", true},
		{"docs/diagram.SVG", true},      // case-insensitive extension
		{"vendor/bundle.min.js", true},  // compound extension
		{"nested/dir/Cargo.lock", true}, // basename match with path
		{"a\\b\\yarn.lock", true},       // windows-style separator
		{"not-a-lock-file.go", false},   // "lock" substring must not match
		{"image.png.go", false},         // extension only at suffix
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := shouldSkip(tt.filename); got != tt.want {
				t.Errorf("shouldSkip(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestBuildDiff(t *testing.T) {
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/octocat/repo/pulls/1/files" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		io.WriteString(w, `[
			{"filename":"main.go","patch":"@@ -1 +1 @@\n-old\n+new","additions":1,"deletions":1},
			{"filename":"go.sum","patch":"@@ hash changes @@","additions":9,"deletions":0},
			{"filename":"empty.go","additions":0,"deletions":0}
		]`)
	}))

	diff, err := BuildDiff(context.Background(), client, PullRequest{Owner: "octocat", Repo: "repo", Number: 1})
	if err != nil {
		t.Fatalf("BuildDiff() error = %v", err)
	}

	if !strings.Contains(diff, "main.go") {
		t.Errorf("diff should include main.go, got:\n%s", diff)
	}
	if strings.Contains(diff, "go.sum") {
		t.Errorf("diff should skip go.sum, got:\n%s", diff)
	}
	if strings.Contains(diff, "empty.go") {
		t.Errorf("diff should skip patch-less empty.go, got:\n%s", diff)
	}
}

func TestPostComment(t *testing.T) {
	var gotBody string
	client := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/repos/octocat/repo/issues/1/comments" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"id":1}`)
	}))

	err := PostComment(context.Background(), client, PullRequest{Owner: "octocat", Repo: "repo", Number: 1}, "Great work.", "claude-sonnet-5")
	if err != nil {
		t.Fatalf("PostComment() error = %v", err)
	}

	if !strings.Contains(gotBody, "Great work.") {
		t.Errorf("comment body missing review text, got: %s", gotBody)
	}
	if !strings.Contains(gotBody, "claude-sonnet-5") {
		t.Errorf("comment body missing model footer, got: %s", gotBody)
	}
}
