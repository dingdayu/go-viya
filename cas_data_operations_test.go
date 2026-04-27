package viya

import (
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestUploadCSVToCASTableFromReaderSendsMultipart(t *testing.T) {
	const csv = "Name,Age\nAlice,13\n"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers/server%201/caslibs/Public%20Data/tables"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got, want := r.FormValue("tableName"), "class table"; got != want {
			t.Fatalf("tableName = %q, want %q", got, want)
		}
		if got, want := r.FormValue("format"), "csv"; got != want {
			t.Fatalf("format = %q, want %q", got, want)
		}
		if got, want := r.FormValue("containsHeaderRow"), "true"; got != want {
			t.Fatalf("containsHeaderRow = %q, want %q", got, want)
		}
		content := readMultipartFile(t, r.MultipartForm.File["file"][0])
		if got := string(content); got != csv {
			t.Fatalf("file content = %q, want %q", got, csv)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"class table","caslibName":"Public Data","rowCount":1,"columnCount":2,"scope":"session"}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL, WithTokenProvider(staticTokenProvider("token-value")))

	table, err := client.UploadCSVToCASTableFromReader(context.Background(), "server 1", "Public Data", "class table", strings.NewReader(csv))
	if err != nil {
		t.Fatalf("UploadCSVToCASTableFromReader() error = %v", err)
	}
	if got, want := table.Name, "class table"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
	if got, want := table.RowCount, int64(1); got != want {
		t.Fatalf("RowCount = %d, want %d", got, want)
	}
}

func TestUploadCSVToCASTableReturnsConflictStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"exists"}`, http.StatusConflict)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	_, err := client.UploadCSVToCASTable(context.Background(), "server-1", "Public", "class", []byte("x\n"))
	if err == nil {
		t.Fatal("UploadCSVToCASTable() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to upload CSV to CAS table, status code: 409") {
		t.Fatalf("error = %q, want 409 context", err.Error())
	}
}

func TestPromoteCASTableSendsScopeBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers/server%201/caslibs/Public%20Data/tables/class%20table"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if got := string(body); !strings.Contains(got, `"scope":"global"`) {
			t.Fatalf("body = %s, want scope global", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"class table","caslibName":"Public Data","scope":"global"}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	table, err := client.PromoteCASTable(context.Background(), "server 1", "Public Data", "class table")
	if err != nil {
		t.Fatalf("PromoteCASTable() error = %v", err)
	}
	if got, want := table.Scope, "global"; got != want {
		t.Fatalf("Scope = %q, want %q", got, want)
	}
}

func TestPromoteCASTableReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"missing"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	_, err := client.PromoteCASTable(context.Background(), "server-1", "Public", "missing")
	if err == nil {
		t.Fatal("PromoteCASTable() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to promote CAS table, status code: 404") {
		t.Fatalf("error = %q, want 404 context", err.Error())
	}
}

func readMultipartFile(t *testing.T, header *multipart.FileHeader) []byte {
	t.Helper()

	file, err := header.Open()
	if err != nil {
		t.Fatalf("open multipart file: %v", err)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("read multipart file: %v", err)
	}
	return content
}
