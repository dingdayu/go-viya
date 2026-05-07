package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetBatchContextByNameSendsNameQuery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/batch/contexts/@item"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("name"), "default context"; got != want {
			t.Fatalf("name query = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"context-1","name":"default context","state":"active","version":1}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(context.Background(), WithBaseURL(u))

	batchContext, err := client.GetBatchContextByName(context.Background(), "default context")
	if err != nil {
		t.Fatalf("GetBatchContextByName() error = %v", err)
	}
	if got, want := batchContext.ID, "context-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
	if got, want := batchContext.Name, "default context"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}
