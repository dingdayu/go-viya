package viya

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadBatchFileFromReaderSendsFileContent(t *testing.T) {
	const body = "data _null_; put 'hello'; run;"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/batch/fileSets/file%20set%201/files/program%20one.sas"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/octet-stream") {
			t.Fatalf("Content-Type = %q, want application/octet-stream", got)
		}

		gotBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if got := string(gotBody); got != body {
			t.Fatalf("body = %q, want %q", got, body)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.UploadBatchFileFromReader(context.Background(), "file set 1", "program one.sas", strings.NewReader(body))
	if err != nil {
		t.Fatalf("UploadBatchFileFromReader() error = %v", err)
	}
}

func TestUploadBatchFileFromReaderReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"missing file set"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.UploadBatchFileFromReader(context.Background(), "file-set-1", "program.sas", strings.NewReader("body"))
	if err == nil {
		t.Fatal("UploadBatchFileFromReader() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "status code: 404") {
		t.Fatalf("error = %q, want status code 404", err.Error())
	}
}
