package viya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoadCASTableToMemorySetsLoadedState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers/server%201/caslibs/Public%20Data/tables/class%20table/state?value=loaded"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got, want := body["outputCaslibName"], any("Public Data"); got != want {
			t.Fatalf("outputCaslibName = %#v, want %#v", got, want)
		}
		if got, want := body["outputTableName"], any("class table"); got != want {
			t.Fatalf("outputTableName = %#v, want %#v", got, want)
		}
		if got, want := body["replace"], any(true); got != want {
			t.Fatalf("replace = %#v, want %#v", got, want)
		}
		if got, want := body["scope"], any("global"); got != want {
			t.Fatalf("scope = %#v, want %#v", got, want)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.LoadCASTableToMemory(context.Background(), "server 1", "Public Data", "class table", true, "global")
	if err != nil {
		t.Fatalf("LoadCASTableToMemory() error = %v", err)
	}
}

func TestUnloadCASTableFromMemorySetsUnloadedState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers/server-1/caslibs/Public/tables/class/state?value=unloaded"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.UnloadCASTableFromMemory(context.Background(), "server-1", "Public", "class")
	if err != nil {
		t.Fatalf("UnloadCASTableFromMemory() error = %v", err)
	}
}

func TestUnloadCASTableFromMemoryReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"conflict"}`, http.StatusConflict)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.UnloadCASTableFromMemory(context.Background(), "server-1", "Public", "class")
	if err == nil {
		t.Fatal("UnloadCASTableFromMemory() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to unload CAS table from memory, status code: 409") {
		t.Fatalf("error = %q, want status code context", err.Error())
	}
}
