package viya

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetConfigurationReturnsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("definitionName"), "test.definition"; got != want {
			t.Fatalf("definitionName = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":1,"items":[{"name":"test"}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	body, err := client.GetConfiguration(t.Context(), "test.definition")
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if !strings.Contains(body, "test.definition") && !strings.Contains(body, "test") {
		t.Fatalf("body = %q, want test definition content", body)
	}
}

func TestGetConfigurationInvalidDefinition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	// GetConfiguration returns the raw response body even for non-200 responses.
	body, err := client.GetConfiguration(t.Context(), "missing")
	if err != nil {
		t.Fatalf("GetConfiguration() error = %v", err)
	}
	if body == "" {
		t.Fatal("GetConfiguration() body = empty, want error body")
	}
}
