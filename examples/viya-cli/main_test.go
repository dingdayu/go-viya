package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCASServersCommandWritesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers?limit=7"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer test-token"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"server-1","name":"cas-shared-default"}]}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLI("cas", "--base-url", server.URL, "--access-token", "test-token", "-o", "json", "servers", "--limit", "7")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}

	var body struct {
		OK   bool `json:"ok"`
		Data struct {
			Count int `json:"count"`
			Items []struct {
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if !body.OK {
		t.Fatal("ok = false, want true")
	}
	if got, want := body.Data.Count, 1; got != want {
		t.Fatalf("count = %d, want %d", got, want)
	}
	if got, want := body.Data.Items[0].Name, "cas-shared-default"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
}

func TestCASCommandMissingFlagWritesFailureJSON(t *testing.T) {
	stdout, _, err := executeCLI("cas", "--base-url", "https://viya.example.com", "--access-token", "test-token", "-o", "json", "tables", "--server", "server-1")
	if err == nil {
		t.Fatal("executeCLI() error = nil, want exit error")
	}

	var body struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if decodeErr := json.Unmarshal([]byte(stdout), &body); decodeErr != nil {
		t.Fatalf("decode stdout: %v", decodeErr)
	}
	if body.OK {
		t.Fatal("ok = true, want false")
	}
	if !strings.Contains(body.Error, "--caslib is required") {
		t.Fatalf("error = %q, want missing caslib", body.Error)
	}
}

func TestCASRowsCommandWritesRowsJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/dataTables/dataSources/cas~fs~cas-shared-default~fs~Public/tables/class/columns":
			_, _ = w.Write([]byte(`{"count":2,"items":[{"name":"Name"},{"name":"Age"}]}`))
		case "/rowSets/tables/cas~fs~cas-shared-default~fs~Public~fs~class/rows":
			if got, want := r.URL.Query().Get("start"), "1"; got != want {
				t.Fatalf("start = %q, want %q", got, want)
			}
			if got, want := r.URL.Query().Get("limit"), "2"; got != want {
				t.Fatalf("limit = %q, want %q", got, want)
			}
			_, _ = w.Write([]byte(`{"count":1,"items":[{"cells":["Alice",13]}]}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	stdout, _, err := executeCLI(
		"cas", "--base-url", server.URL, "--access-token", "test-token",
		"-o", "json",
		"rows", "--server", "cas-shared-default", "--caslib", "Public", "--table", "class", "--start", "1", "--limit", "2",
	)
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}

	var body struct {
		OK   bool `json:"ok"`
		Data struct {
			Columns []string         `json:"columns"`
			Rows    []map[string]any `json:"rows"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(stdout), &body); err != nil {
		t.Fatalf("decode stdout: %v", err)
	}
	if !body.OK {
		t.Fatal("ok = false, want true")
	}
	if got, want := body.Data.Columns[0], "Name"; got != want {
		t.Fatalf("first column = %q, want %q", got, want)
	}
	if got, want := body.Data.Rows[0]["Name"], any("Alice"); got != want {
		t.Fatalf("Name = %#v, want %#v", got, want)
	}
}

func TestCASCommandLoadsConfigFromEnvironment(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":0,"items":[]}`))
	}))
	defer server.Close()

	t.Setenv("VIYA_BASE_URL", server.URL)
	t.Setenv("VIYA_ACCESS_TOKEN", "env-token")

	stdout, _, err := executeCLI("cas", "-o", "json", "servers")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, `"ok": true`) {
		t.Fatalf("stdout = %s, want ok true", stdout)
	}
}

func TestCASServersCommandDefaultsToText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"server-1","name":"cas-shared-default","description":"shared"}]}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLI("cas", "--base-url", server.URL, "--access-token", "test-token", "servers")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, "NAME") || !strings.Contains(stdout, "cas-shared-default") {
		t.Fatalf("stdout = %s, want text table", stdout)
	}
}

func executeCLI(args ...string) (stdout string, stderr string, err error) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := newRootCommand(cliIO{
		stdout: &out,
		stderr: &errOut,
		stdin:  strings.NewReader(""),
	})
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}
