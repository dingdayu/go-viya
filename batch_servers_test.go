package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetBatchServersListDecodesServers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/batch/servers"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"server-1","contextId":"context-1","hostName":"host","state":"available","isPermanent":true,"version":1}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	servers, err := client.GetBatchServersList(context.Background())
	if err != nil {
		t.Fatalf("GetBatchServersList() error = %v", err)
	}
	if got, want := servers.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if len(servers.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(servers.Items))
	}
	if got, want := servers.Items[0].State, "available"; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
	if got, want := servers.Items[0].HostName, "host"; got != want {
		t.Fatalf("HostName = %q, want %q", got, want)
	}
}

func TestGetBatchServerInfoDecodesServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/batch/servers/server%201"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"server 1","contextId":"context-1","hostName":"host","processId":"proc-1","queueName":"default","state":"busy","version":1}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	serverInfo, err := client.GetBatchServerInfo(context.Background(), "server 1")
	if err != nil {
		t.Fatalf("GetBatchServerInfo() error = %v", err)
	}
	if got, want := serverInfo.ID, "server 1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
	if got, want := serverInfo.ProcessID, "proc-1"; got != want {
		t.Fatalf("ProcessID = %q, want %q", got, want)
	}
}

func TestDeleteBatchServerSendsDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodDelete; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/batch/servers/server%201"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.DeleteBatchServer(context.Background(), "server 1")
	if err != nil {
		t.Fatalf("DeleteBatchServer() error = %v", err)
	}
}
