package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetComputeJobLogDecodesLines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/compute/sessions/session-1/jobs/0/log"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/json, application/vnd.sas.collection+json;version=2, application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":2,"items":[{"line":"NOTE: start","type":"note"},{"line":"run;","type":"source"}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	log, err := client.GetComputeJobLog(context.Background(), "session-1", "0")
	if err != nil {
		t.Fatalf("GetComputeJobLog() error = %v", err)
	}
	if got, want := len(log.Items), 2; got != want {
		t.Fatalf("len(Items) = %d, want %d", got, want)
	}
	if got, want := log.Items[0].Type, "note"; got != want {
		t.Fatalf("Type = %q, want %q", got, want)
	}
}

func TestGetComputeJobListingTextReturnsPlainText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/compute/sessions/session-1/jobs/0/listing"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "text/plain, application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("listing line\n"))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	text, err := client.GetComputeJobListingText(context.Background(), "session-1", "0")
	if err != nil {
		t.Fatalf("GetComputeJobListingText() error = %v", err)
	}
	if got, want := text, "listing line"; got != want {
		t.Fatalf("text = %q, want %q", got, want)
	}
}
