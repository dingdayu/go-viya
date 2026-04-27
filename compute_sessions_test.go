package viya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCreateComputeSessionSendsRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/compute/contexts/context%201/sessions"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var req CreateComputeSessionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if got, want := req.Name, "session one"; got != want {
			t.Fatalf("Name = %q, want %q", got, want)
		}
		if req.Environment == nil || len(req.Environment.Options) != 1 || req.Environment.Options[0] != "fullstimer" {
			t.Fatalf("Environment = %#v, want fullstimer option", req.Environment)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"session-1","name":"session one","state":"idle","version":2}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	session, err := client.CreateComputeSession(context.Background(), "context 1", CreateComputeSessionRequest{
		Name:        "session one",
		Environment: &ComputeEnvironment{Options: []string{"fullstimer"}},
	})
	if err != nil {
		t.Fatalf("CreateComputeSession() error = %v", err)
	}
	if got, want := session.ID, "session-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestGetComputeSessionStateReturnsPlainTextState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/compute/sessions/session-1/state"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "text/plain, application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("idle\n"))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	state, err := client.GetComputeSessionState(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetComputeSessionState() error = %v", err)
	}
	if got, want := state, "idle"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestCancelComputeSessionSendsIfMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/compute/sessions/session-1/state"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("value"), "canceled"; got != want {
			t.Fatalf("value query = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("If-Match"), `"etag-1"`; got != want {
			t.Fatalf("If-Match = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("canceled\n"))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	state, err := client.CancelComputeSession(context.Background(), "session-1", `"etag-1"`)
	if err != nil {
		t.Fatalf("CancelComputeSession() error = %v", err)
	}
	if got, want := state, "canceled"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestDeleteComputeSessionAcceptsAcceptedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodDelete; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/compute/sessions/session-1"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	if err := client.DeleteComputeSession(context.Background(), "session-1"); err != nil {
		t.Fatalf("DeleteComputeSession() error = %v", err)
	}
}
