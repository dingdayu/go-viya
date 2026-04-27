package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
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

	client := NewClient(context.Background(), server.URL)

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
