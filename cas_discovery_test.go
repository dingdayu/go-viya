package viya

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestGetCASServersRequestsCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/casManagement/servers"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("limit"), "10"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"server-1","name":"cas-shared-default","description":"shared"}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL, WithTokenProvider(staticTokenProvider("token-value")))

	servers, err := client.GetCASServers(context.Background(), ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("GetCASServers() error = %v", err)
	}
	if got, want := servers.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if got, want := servers.Items[0].Name, "cas-shared-default"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestGetCASTablesEscapesPathAndSetsPaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/casManagement/servers/server%201/caslibs/Public%20Data/tables?limit=25&start=5"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"name":"class","rowCount":19,"columnCount":5}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	tables, err := client.GetCASTables(context.Background(), "server 1", "Public Data", ListOptions{Start: 5, Limit: 25})
	if err != nil {
		t.Fatalf("GetCASTables() error = %v", err)
	}
	if got, want := tables.Items[0].RowCount, int64(19); got != want {
		t.Fatalf("RowCount = %d, want %d", got, want)
	}
}

func TestGetCASTableColumnsReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"missing"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	_, err := client.GetCASTableColumns(context.Background(), "server-1", "Public", "missing", ListOptions{})
	if err == nil {
		t.Fatal("GetCASTableColumns() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to get CAS table columns, status code: 404") {
		t.Fatalf("error = %q, want status code context", err.Error())
	}
}

func TestGetCASTableRowsUsesDataTablesAndRowSets(t *testing.T) {
	var requests []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, r.RequestURI)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.EscapedPath() {
		case "/dataTables/dataSources/cas~fs~server%201~fs~Public%20Data/tables/class%20table/columns":
			if got, want := r.URL.Query().Get("limit"), "1000"; got != want {
				t.Fatalf("column limit = %q, want %q", got, want)
			}
			_, _ = w.Write([]byte(`{"count":2,"items":[{"name":"Name","type":"string","index":0},{"name":"Age","type":"double","index":1}]}`))
		case "/rowSets/tables/cas~fs~server%201~fs~Public%20Data~fs~class%20table/rows":
			if got, want := r.URL.Query().Get("start"), "2"; got != want {
				t.Fatalf("row start = %q, want %q", got, want)
			}
			if got, want := r.URL.Query().Get("limit"), "3"; got != want {
				t.Fatalf("row limit = %q, want %q", got, want)
			}
			_, _ = w.Write([]byte(`{"count":1,"items":[{"cells":["Alice",13]}]}`))
		default:
			t.Fatalf("unexpected path %q", r.URL.EscapedPath())
		}
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	rows, err := client.GetCASTableRows(context.Background(), "server 1", "Public Data", "class table", ListOptions{Start: 2, Limit: 3})
	if err != nil {
		t.Fatalf("GetCASTableRows() error = %v", err)
	}
	if !reflect.DeepEqual(rows.Columns, []string{"Name", "Age"}) {
		t.Fatalf("Columns = %#v, want Name/Age", rows.Columns)
	}
	if got, want := rows.Rows[0]["Name"], any("Alice"); got != want {
		t.Fatalf("Name cell = %#v, want %#v", got, want)
	}
	if got, want := rows.Rows[0]["Age"], any(float64(13)); got != want {
		t.Fatalf("Age cell = %#v, want %#v", got, want)
	}
	if got, want := len(requests), 2; got != want {
		t.Fatalf("requests = %d, want %d", got, want)
	}
}
