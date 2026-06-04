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
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.RequestURI)
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		switch r.RequestURI {
		case "/reports/reports/report%201":
			_, _ = w.Write([]byte(`{"id":"report 1","name":"My Report"}`))
		case "/reports/report%201/content":
			_, _ = w.Write([]byte(`{"layout":"single","sections":[{"name":"Overview"}]}`))
		default:
			t.Fatalf("request URI = %q, want report metadata or content endpoint", r.RequestURI)
		}
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
	definition, ok := reportDef["definition"].(map[string]any)
	if !ok {
		t.Fatalf("definition = %#v, want map[string]any", reportDef["definition"])
	}
	if got, want := definition["layout"], "single"; got != want {
		t.Fatalf("definition.layout = %q, want %q", got, want)
	}
	if got, want := calls, []string{"GET /reports/reports/report%201", "GET /reports/report%201/content"}; !stringSlicesEqual(got, want) {
		t.Fatalf("calls = %v, want %v", got, want)
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
		if got, want := body["reportUri"], "/reports/reports/folder%2Freport%201"; got != want {
			t.Fatalf("reportUri = %q, want %q", got, want)
		}
		if got, want := body["layoutType"], "thumbnail"; got != want {
			t.Fatalf("layoutType = %q, want %q", got, want)
		}
		if got, want := body["sectionIndex"], float64(2); got != want {
			t.Fatalf("sectionIndex = %v, want %v", got, want)
		}
		if got, want := body["renderLimit"], float64(-1); got != want {
			t.Fatalf("renderLimit = %v, want %v", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"job-1","state":"completed","images":[{"sectionIndex":2,"sectionName":"vi6","sectionLabel":"Page 1","elementName":"ve41","visualType":"Table","size":"800x600","state":"completed","links":[{"method":"GET","rel":"image","href":"/reportImages/images/image-1.svg","uri":"/reportImages/images/image-1.svg","type":"image/svg+xml"}]}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	job, err := client.GetReportImage(t.Context(), "folder/report 1", ReportImageOptions{SectionIndex: 2})
	if err != nil {
		t.Fatalf("GetReportImage() error = %v", err)
	}
	if got, want := job.ID, "job-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
	if got, want := len(job.Images), 1; got != want {
		t.Fatalf("len(Images) = %d, want %d", got, want)
	}
	if got, want := job.Images[0].SectionIndex, 2; got != want {
		t.Fatalf("Images[0].SectionIndex = %d, want %d", got, want)
	}
	if got, want := job.Images[0].SectionLabel, "Page 1"; got != want {
		t.Fatalf("Images[0].SectionLabel = %q, want %q", got, want)
	}
	if got, want := job.Images[0].Links[0].Rel, "image"; got != want {
		t.Fatalf("Images[0].Links[0].Rel = %q, want %q", got, want)
	}
	if got, want := job.Images[0].Links[0].Href, "/reportImages/images/image-1.svg"; got != want {
		t.Fatalf("Images[0].Links[0].Href = %q, want %q", got, want)
	}
}

func TestGetReportImagePreservesNoLimitRenderLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if got, want := body["renderLimit"], float64(-1); got != want {
			t.Fatalf("renderLimit = %v, want %v", got, want)
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

	_, err = client.GetReportImage(t.Context(), "report-1", ReportImageOptions{RenderLimit: -1})
	if err != nil {
		t.Fatalf("GetReportImage() error = %v", err)
	}
}

func TestGetReportImageRejectsEmptyReportID(t *testing.T) {
	client := NewClient(t.Context(), WithBaseURL(&url.URL{}), WithTokenProvider(staticTokenProvider("token")))

	_, err := client.GetReportImage(t.Context(), "", ReportImageOptions{})
	if err == nil {
		t.Fatal("expected error for empty reportID")
	}
}

func stringSlicesEqual(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
