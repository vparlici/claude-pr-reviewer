package claude

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// stubEndpoint points the package endpoint at srv for the duration of the test.
func stubEndpoint(t *testing.T, srv *httptest.Server) {
	t.Helper()
	original := endpoint
	endpoint = srv.URL
	t.Cleanup(func() { endpoint = original })
}

func TestReviewSuccess(t *testing.T) {
	var gotBody request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("x-api-key"); got != "secret-key" {
			t.Errorf("x-api-key = %q, want %q", got, "secret-key")
		}
		if got := r.Header.Get("anthropic-version"); got != version {
			t.Errorf("anthropic-version = %q, want %q", got, version)
		}
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &gotBody); err != nil {
			t.Fatalf("decoding request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"content":[{"type":"text","text":"  Looks good.  "}]}`)
	}))
	defer srv.Close()
	stubEndpoint(t, srv)

	got, err := Review(context.Background(), "secret-key", "the diff")
	if err != nil {
		t.Fatalf("Review() error = %v", err)
	}
	if got != "Looks good." {
		t.Errorf("Review() = %q, want trimmed %q", got, "Looks good.")
	}

	if gotBody.Model != Model {
		t.Errorf("request Model = %q, want %q", gotBody.Model, Model)
	}
	if gotBody.MaxTokens != maxTokens {
		t.Errorf("request MaxTokens = %d, want %d", gotBody.MaxTokens, maxTokens)
	}
	if len(gotBody.Messages) != 1 || !strings.Contains(gotBody.Messages[0].Content, "the diff") {
		t.Errorf("request Messages = %+v, want one message containing the diff", gotBody.Messages)
	}
}

func TestReviewAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":{"type":"invalid_request_error","message":"bad model"}}`)
	}))
	defer srv.Close()
	stubEndpoint(t, srv)

	_, err := Review(context.Background(), "k", "diff")
	if err == nil {
		t.Fatal("Review() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "bad model") {
		t.Errorf("error = %v, want it to mention the API message", err)
	}
}

func TestReviewEmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `{"content":[]}`)
	}))
	defer srv.Close()
	stubEndpoint(t, srv)

	_, err := Review(context.Background(), "k", "diff")
	if err == nil {
		t.Fatal("Review() error = nil, want error for empty review")
	}
}

func TestReviewMalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, `not json`)
	}))
	defer srv.Close()
	stubEndpoint(t, srv)

	_, err := Review(context.Background(), "k", "diff")
	if err == nil {
		t.Fatal("Review() error = nil, want error for malformed JSON")
	}
}
