package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetComputeContextsRequestsFullItems(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/compute/contexts"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept-Item"), "application/vnd.sas.compute.context+json"; got != want {
			t.Fatalf("Accept-Item = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"context-1","name":"default","version":1}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	contexts, err := client.GetComputeContexts(context.Background())
	if err != nil {
		t.Fatalf("GetComputeContexts() error = %v", err)
	}
	if got, want := contexts.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if got, want := contexts.Items[0].Name, "default"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestGetComputeContextInfoEscapesContextID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/compute/contexts/context%201"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"context 1","name":"default","launchType":"service","version":1}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	contextInfo, err := client.GetComputeContextInfo(context.Background(), "context 1")
	if err != nil {
		t.Fatalf("GetComputeContextInfo() error = %v", err)
	}
	if got, want := contextInfo.ID, "context 1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}
