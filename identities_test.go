package viya

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetIdentitiesUsersReturnsUsers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/identities/users?limit=100"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"user-1","name":"Alice","providerId":"ldap","type":"user","state":"active"}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	users, err := client.GetIdentitiesUsers(t.Context())
	if err != nil {
		t.Fatalf("GetIdentitiesUsers() error = %v", err)
	}
	if got, want := users.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if len(users.Items) != 1 {
		t.Fatalf("len(Items) = %d, want 1", len(users.Items))
	}
	if got, want := users.Items[0].Name, "Alice"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestGetIdentitiesUsersReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"unauthorized"}`, http.StatusUnauthorized)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	_, err = client.GetIdentitiesUsers(t.Context())
	if err == nil {
		t.Fatal("GetIdentitiesUsers() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "status code: 401") {
		t.Fatalf("error = %q, want 401", err.Error())
	}
}
