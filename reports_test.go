package viya

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetReportsRequestsCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/reports/reports"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("limit"), "10"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"report-1","name":"My Report","creationTimeStamp":"2026-05-29T08:00:00Z"}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	reports, err := client.GetReports(t.Context(), ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("GetReports() error = %v", err)
	}
	if got, want := reports.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if got, want := reports.Items[0].Name, "My Report"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
	if reports.Items[0].CreationTimeStamp.IsZero() {
		t.Fatal("CreationTimeStamp was not parsed")
	}
}

func TestGetReportEscapesIDAndReturnsDefinition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/reports/reports/report%201"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"report 1","name":"My Report","definition":{"layout":"single"}}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	reportDef, err := client.GetReport(t.Context(), "report 1")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if got, want := reportDef["name"], "My Report"; got != want {
		t.Fatalf("name = %q, want %q", got, want)
	}
}

func TestGetReportRejectsEmptyReportID(t *testing.T) {
	client := NewClient(t.Context(), WithBaseURL(&url.URL{}), WithTokenProvider(staticTokenProvider("token")))

	_, err := client.GetReport(t.Context(), "")
	if err == nil {
		t.Fatal("expected error for empty reportID")
	}
	if got, want := err.Error(), "invalid parameter reportID: must not be empty"; got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestGetReportImageCreatesJob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/reportImages/jobs"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/vnd.sas.report.images.job+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Type"), "application/vnd.sas.report.images.job.request+json"; got != want {
			t.Fatalf("Content-Type = %q, want %q", got, want)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if got, want := body["reportUri"], "/reports/reports/report-1"; got != want {
			t.Fatalf("reportUri = %q, want %q", got, want)
		}
		if got, want := body["layoutType"], "thumbnail"; got != want {
			t.Fatalf("layoutType = %q, want %q", got, want)
		}
		if got, want := body["sectionIndex"], float64(2); got != want {
			t.Fatalf("sectionIndex = %v, want %v", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"job-1","state":"running"}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	job, err := client.GetReportImage(t.Context(), "report-1", ReportImageOptions{SectionIndex: 2})
	if err != nil {
		t.Fatalf("GetReportImage() error = %v", err)
	}
	if got, want := job.ID, "job-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestGetReportImageRejectsEmptyReportID(t *testing.T) {
	client := NewClient(t.Context(), WithBaseURL(&url.URL{}), WithTokenProvider(staticTokenProvider("token")))

	_, err := client.GetReportImage(t.Context(), "", ReportImageOptions{})
	if err == nil {
		t.Fatal("expected error for empty reportID")
	}
}
