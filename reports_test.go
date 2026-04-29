package viya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetReportsSetsPagingFilterAndAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/reports/reports"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("limit"), "7"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("filter"), "contains(name,'sales')"; got != want {
			t.Fatalf("filter = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"report-1","name":"Sales","description":"Quarterly","createdBy":"user1"}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL, WithTokenProvider(staticTokenProvider("token-value")))

	reports, err := client.GetReports(context.Background(), ReportListOptions{Limit: 7, FilterName: "sales"})
	if err != nil {
		t.Fatalf("GetReports() error = %v", err)
	}
	if got, want := reports.Items[0].Name, "Sales"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestGetReportEscapesIDAndDecodesDefinition(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/reports/reports/report%201"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"report 1","name":"Sales","definition":{"pages":[{"name":"Page 1"}]}}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	report, err := client.GetReport(context.Background(), "report 1")
	if err != nil {
		t.Fatalf("GetReport() error = %v", err)
	}
	if got, want := report.Definition["pages"].([]any)[0].(map[string]any)["name"], any("Page 1"); got != want {
		t.Fatalf("definition page name = %#v, want %#v", got, want)
	}
}

func TestGetReportReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"missing"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	_, err := client.GetReport(context.Background(), "missing")
	if err == nil {
		t.Fatal("GetReport() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "failed to get report, status code: 404") {
		t.Fatalf("error = %q, want status context", err.Error())
	}
}

func TestCreateReportImageJobSendsDefaults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/reportImages/jobs"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Type"), "application/vnd.sas.report.images.job.request+json"; !strings.HasPrefix(got, want) {
			t.Fatalf("Content-Type = %q, want %q", got, want)
		}
		var body struct {
			ReportURI     string `json:"reportUri"`
			LayoutType    string `json:"layoutType"`
			SelectionType string `json:"selectionType"`
			SectionIndex  int    `json:"sectionIndex"`
			Size          string `json:"size"`
			RenderLimit   int    `json:"renderLimit"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got, want := body.ReportURI, "/reports/reports/report%201"; got != want {
			t.Fatalf("reportUri = %q, want %q", got, want)
		}
		if got, want := body.LayoutType, "thumbnail"; got != want {
			t.Fatalf("layoutType = %q, want %q", got, want)
		}
		if got, want := body.SelectionType, "perSection"; got != want {
			t.Fatalf("selectionType = %q, want %q", got, want)
		}
		if got, want := body.SectionIndex, 2; got != want {
			t.Fatalf("sectionIndex = %d, want %d", got, want)
		}
		if got, want := body.Size, "800x600"; got != want {
			t.Fatalf("size = %q, want %q", got, want)
		}
		if got, want := body.RenderLimit, 1; got != want {
			t.Fatalf("renderLimit = %d, want %d", got, want)
		}
		w.Header().Set("Content-Type", "application/vnd.sas.report.images.job+json")
		_, _ = w.Write([]byte(`{"id":"image-job-1","state":"running","reportUri":"/reports/reports/report%201"}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	job, err := client.CreateReportImageJob(context.Background(), ReportImageJobRequest{ReportID: "report 1", SectionIndex: 2})
	if err != nil {
		t.Fatalf("CreateReportImageJob() error = %v", err)
	}
	if got, want := job.ID, "image-job-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestCreateVisualAnalyticsReportSendsOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/visualAnalytics/reports"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Type"), "application/vnd.sas.visual.analytics.reports.request+json"; !strings.HasPrefix(got, want) {
			t.Fatalf("Content-Type = %q, want %q", got, want)
		}
		var body VisualAnalyticsReportRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got, want := body.Version, 1; got != want {
			t.Fatalf("version = %d, want %d", got, want)
		}
		if got, want := body.ResultFolder, "/folders/folders/@myFolder"; got != want {
			t.Fatalf("resultFolder = %q, want %q", got, want)
		}
		if got, want := body.ResultReportName, "Sales Dashboard"; got != want {
			t.Fatalf("resultReportName = %q, want %q", got, want)
		}
		if _, ok := body.Operations[0]["addData"]; !ok {
			t.Fatalf("first operation = %#v, want addData", body.Operations[0])
		}
		w.Header().Set("Content-Type", "application/vnd.sas.report.operations.results+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"resultReportId":"report-1","resultReportName":"Sales Dashboard","resultReportUri":"/reports/reports/report-1","status":"Success"}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	result, err := client.CreateVisualAnalyticsReport(context.Background(), VisualAnalyticsReportRequest{
		ResultFolder:       "/folders/folders/@myFolder",
		ResultReportName:   "Sales Dashboard",
		ResultNameConflict: "replace",
		Operations: []map[string]any{
			{"addData": map[string]any{"cas": map[string]any{"server": "cas-shared-default", "library": "Public", "table": "SALES"}}},
		},
	})
	if err != nil {
		t.Fatalf("CreateVisualAnalyticsReport() error = %v", err)
	}
	if got, want := result.ResultReportID, "report-1"; got != want {
		t.Fatalf("ResultReportID = %q, want %q", got, want)
	}
}

func TestCreateDashboardBuildsAgentFriendlyOperations(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body VisualAnalyticsReportRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got, want := len(body.Operations), 3; got != want {
			t.Fatalf("operations = %d, want %d: %#v", got, want, body.Operations)
		}
		addData := body.Operations[0]["addData"].(map[string]any)
		cas := addData["cas"].(map[string]any)
		if got, want := cas["table"], any("SALES"); got != want {
			t.Fatalf("table = %#v, want %#v", got, want)
		}
		if _, ok := body.Operations[1]["addPage"]; !ok {
			t.Fatalf("second operation = %#v, want addPage", body.Operations[1])
		}
		addObject := body.Operations[2]["addObject"].(map[string]any)
		object := addObject["object"].(map[string]any)
		bar := object["barChart"].(map[string]any)
		roles := bar["dataRoles"].(map[string]any)
		if got, want := roles["category"], any("REGION"); got != want {
			t.Fatalf("category = %#v, want %#v", got, want)
		}
		w.Header().Set("Content-Type", "application/vnd.sas.report.operations.results+json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"resultReportId":"report-1","status":"Success"}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	_, err := client.CreateDashboard(context.Background(), CreateDashboardRequest{
		Name:       "Sales Dashboard",
		FolderURI:  "/folders/folders/@myFolder",
		ServerID:   "cas-shared-default",
		CaslibName: "Public",
		TableName:  "SALES",
		Spec: DashboardSpec{Pages: []DashboardPageSpec{{
			Name: "Overview",
			Objects: []DashboardObjectSpec{{
				Type:     "barChart",
				Title:    "Sales by Region",
				Category: "REGION",
				Measure:  "SALES",
			}},
		}}},
	})
	if err != nil {
		t.Fatalf("CreateDashboard() error = %v", err)
	}
}
