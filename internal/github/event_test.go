package github

import (
	"os"
	"path/filepath"
	"testing"
)

// writeEvent writes content to a temp file and returns its path.
func writeEvent(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "event.json")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("writing event file: %v", err)
	}
	return path
}

func TestParseEvent(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    PullRequest
		wantErr bool
	}{
		{
			name:    "top-level number",
			payload: `{"number":42,"repository":{"name":"repo","owner":{"login":"octocat"}}}`,
			want:    PullRequest{Owner: "octocat", Repo: "repo", Number: 42},
		},
		{
			name:    "falls back to pull_request.number",
			payload: `{"pull_request":{"number":7},"repository":{"name":"repo","owner":{"login":"octocat"}}}`,
			want:    PullRequest{Owner: "octocat", Repo: "repo", Number: 7},
		},
		{
			name:    "missing number",
			payload: `{"repository":{"name":"repo","owner":{"login":"octocat"}}}`,
			wantErr: true,
		},
		{
			name:    "missing repository",
			payload: `{"number":42}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			payload: `{not json`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEvent(writeEvent(t, tt.payload))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseEvent() error = nil, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseEvent() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("ParseEvent() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestParseEventMissingFile(t *testing.T) {
	if _, err := ParseEvent(filepath.Join(t.TempDir(), "does-not-exist.json")); err == nil {
		t.Fatal("ParseEvent() error = nil, want error for missing file")
	}
}
