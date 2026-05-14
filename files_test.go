package viya

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetFilesSetsPagingFilterAndAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/files/files"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("start"), "2"; got != want {
			t.Fatalf("start = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("limit"), "7"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("filter"), "contains(name,'report')"; got != want {
			t.Fatalf("filter = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"file-1","name":"report.txt","contentType":"text/plain","size":12}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	files, err := client.GetFiles(t.Context(), FileListOptions{Start: 2, Limit: 7, FilterName: "report"})
	if err != nil {
		t.Fatalf("GetFiles() error = %v", err)
	}
	if got, want := files.Items[0].Name, "report.txt"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestUploadFileFromReaderSendsHeadersAndBody(t *testing.T) {
	const body = "hello file"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/files/files"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Disposition"), `attachment; filename="report.txt"`; got != want {
			t.Fatalf("Content-Disposition = %q, want %q", got, want)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "text/plain") {
			t.Fatalf("Content-Type = %q, want text/plain", got)
		}
		gotBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if got := string(gotBody); got != body {
			t.Fatalf("body = %q, want %q", got, body)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"file-1","name":"report.txt","contentType":"text/plain","size":10}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	file, err := client.UploadFileFromReader(t.Context(), "report.txt", "text/plain", strings.NewReader(body))
	if err != nil {
		t.Fatalf("UploadFileFromReader() error = %v", err)
	}
	if got, want := file.ID, "file-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestDownloadFileEscapesIDAndReturnsBytes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/files/files/file%201/content"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		_, _ = w.Write([]byte("downloaded"))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	content, err := client.DownloadFile(t.Context(), "file 1")
	if err != nil {
		t.Fatalf("DownloadFile() error = %v", err)
	}
	if got, want := string(content), "downloaded"; got != want {
		t.Fatalf("content = %q, want %q", got, want)
	}
}

func TestUploadFileSendsContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Type"), "text/plain"; got != want {
			t.Fatalf("Content-Type = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"file-1","name":"test.txt"}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	file, err := client.UploadFile(t.Context(), "test.txt", "text/plain", []byte("content"))
	if err != nil {
		t.Fatalf("UploadFile() error = %v", err)
	}
	if got, want := file.Name, "test.txt"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestDownloadFileReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"missing"}`, http.StatusNotFound)
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u))

	_, err = client.DownloadFile(t.Context(), "missing")
	if err == nil {
		t.Fatal("DownloadFile() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to download file, status code: 404") {
		t.Fatalf("error = %q, want 404 context", err.Error())
	}
}
