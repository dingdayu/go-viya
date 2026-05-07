package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type staticTokenProvider string

func (p staticTokenProvider) Token(context.Context) (string, error) {
	return string(p), nil
}

func TestNewClientSetsBearerAuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(context.Background(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	resp, err := client.client.R().SetContext(context.Background()).Get("/")
	if err != nil {
		t.Fatalf("GET error = %v", err)
	}
	if resp.StatusCode() != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", resp.StatusCode(), http.StatusNoContent)
	}
}
