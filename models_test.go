package viya

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestGetMLProjectsRequestsCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/mlPipelineAutomation/projects"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("limit"), "10"; got != want {
			t.Fatalf("limit = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"project-1","name":"Churn Model","state":"completed","creationTimeStamp":"2026-05-29T08:00:00Z"}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	projects, err := client.GetMLProjects(t.Context(), ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("GetMLProjects() error = %v", err)
	}
	if got, want := projects.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if got, want := projects.Items[0].State, "completed"; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
	if projects.Items[0].CreationTimeStamp.IsZero() {
		t.Fatal("CreationTimeStamp was not parsed")
	}
}

func TestNewCreateMLProjectRequestDefaults(t *testing.T) {
	req := NewCreateMLProjectRequest("Churn", "/dataTables/dataSources/cas~fs~cas/tables/HMEQ", "BAD")

	if got, want := req.Type, "predictive"; got != want {
		t.Fatalf("Type = %q, want %q", got, want)
	}
	if got, want := req.PipelineBuildMethod, "automatic"; got != want {
		t.Fatalf("PipelineBuildMethod = %q, want %q", got, want)
	}
	if !req.Settings.ApplyGlobalMetadata || !req.Settings.AutoRun || req.Settings.NumberOfModels != 5 {
		t.Fatalf("Settings = %#v, want SAS MCP defaults", req.Settings)
	}
	if got, want := req.AnalyticsProjectAttributes.ClassSelectionStatistic, "ks"; got != want {
		t.Fatalf("ClassSelectionStatistic = %q, want %q", got, want)
	}
}

func TestCreateMLProjectSendsSASMCPPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/mlPipelineAutomation/projects"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if got, want := body["pipelineBuildMethod"], "automatic"; got != want {
			t.Fatalf("pipelineBuildMethod = %q, want %q", got, want)
		}
		attrs := body["analyticsProjectAttributes"].(map[string]any)
		if got, want := attrs["targetVariable"], "BAD"; got != want {
			t.Fatalf("targetVariable = %q, want %q", got, want)
		}
		settings := body["settings"].(map[string]any)
		if got, want := settings["autoRun"], true; got != want {
			t.Fatalf("autoRun = %v, want %v", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"project-1","name":"Churn"}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	project, err := client.CreateMLProject(t.Context(), NewCreateMLProjectRequest("Churn", "/data", "BAD"))
	if err != nil {
		t.Fatalf("CreateMLProject() error = %v", err)
	}
	if got, want := project.ID, "project-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestCreateMLProjectRejectsMissingRequiredFields(t *testing.T) {
	client := NewClient(t.Context(), WithBaseURL(&url.URL{}), WithTokenProvider(staticTokenProvider("token")))

	_, err := client.CreateMLProject(t.Context(), CreateMLProjectRequest{})
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestRunMLProjectGetsAndPutsWithRetrainAction(t *testing.T) {
	var calls []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls = append(calls, r.Method+" "+r.RequestURI)
		if got, want := r.URL.EscapedPath(), "/mlPipelineAutomation/projects/project%201"; got != want {
			t.Fatalf("escaped path = %q, want %q", got, want)
		}

		switch r.Method {
		case http.MethodGet:
			if got, want := r.Header.Get("Accept"), "application/vnd.sas.analytics.ml.pipeline.automation.project+json"; got != want {
				t.Fatalf("GET Accept = %q, want %q", got, want)
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Etag", `"abc"`)
			_, _ = w.Write([]byte(`{"id":"project 1","name":"Churn","dataTableUri":"/dataTables/dataSources/cas~fs~cas/tables/HMEQ","settings":{"autoRun":true,"numberOfModels":5},"analyticsProjectAttributes":{"targetVariable":"BAD","targetLevel":"binary"}}`))
		case http.MethodPut:
			if got, want := r.URL.Query().Get("action"), "retrainProject"; got != want {
				t.Fatalf("action = %q, want %q", got, want)
			}
			if got, want := r.Header.Get("If-Match"), `"abc"`; got != want {
				t.Fatalf("If-Match = %q, want %q", got, want)
			}
			if got, want := r.Header.Get("Accept-Language"), "en"; got != want {
				t.Fatalf("Accept-Language = %q, want %q", got, want)
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			if got, want := body["dataTableUri"], "/dataTables/dataSources/cas~fs~cas/tables/HMEQ"; got != want {
				t.Fatalf("dataTableUri = %q, want %q", got, want)
			}
			if _, ok := body["settings"].(map[string]any); !ok {
				t.Fatalf("settings = %#v, want object", body["settings"])
			}
			if _, ok := body["analyticsProjectAttributes"].(map[string]any); !ok {
				t.Fatalf("analyticsProjectAttributes = %#v, want object", body["analyticsProjectAttributes"])
			}
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("unexpected method %q", r.Method)
		}
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	project, err := client.RunMLProject(t.Context(), "project 1")
	if err != nil {
		t.Fatalf("RunMLProject() error = %v", err)
	}
	if got, want := project.ID, "project 1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
	if got, want := len(calls), 2; got != want {
		t.Fatalf("calls = %d, want %d", got, want)
	}
}

func TestGetRegisteredModelsRequestsCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/modelRepository/models"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"model-1","name":"Gradient Boost Model","modelVersionName":"v1"}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	models, err := client.GetRegisteredModels(t.Context(), ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("GetRegisteredModels() error = %v", err)
	}
	if got, want := models.Items[0].ModelVersionName, "v1"; got != want {
		t.Fatalf("ModelVersionName = %q, want %q", got, want)
	}
}

func TestGetModelsAndDecisionsRequestsMASModules(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/microanalyticScore/modules"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"mas-1","name":"Score Module"}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	modules, err := client.GetModelsAndDecisions(t.Context(), ListOptions{Limit: 10})
	if err != nil {
		t.Fatalf("GetModelsAndDecisions() error = %v", err)
	}
	if got, want := modules.Items[0].Name, "Score Module"; got != want {
		t.Fatalf("Name = %q, want %q", got, want)
	}
}

func TestNewScoreDataRequestPreservesInputOrder(t *testing.T) {
	req := NewScoreDataRequest(
		ScoreInput{Name: "income", Value: 50000},
		ScoreInput{Name: "age", Value: 35},
	)

	if got, want := len(req.Inputs), 2; got != want {
		t.Fatalf("len(Inputs) = %d, want %d", got, want)
	}
	if got, want := req.Inputs[0].Name, "income"; got != want {
		t.Fatalf("Inputs[0].Name = %q, want %q", got, want)
	}
	if got, want := req.Inputs[1].Name, "age"; got != want {
		t.Fatalf("Inputs[1].Name = %q, want %q", got, want)
	}
}

func TestScoreDataSendsMASStepPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/microanalyticScore/modules/mod%201/steps/score%20step"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		var body ScoreDataRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if got, want := len(body.Inputs), 1; got != want {
			t.Fatalf("len(Inputs) = %d, want %d", got, want)
		}
		if got, want := body.Inputs[0].Name, "age"; got != want {
			t.Fatalf("input name = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"outputs":[{"name":"P_BAD1","value":0.42}]}`))
	}))
	defer server.Close()

	u, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("url.Parse() error = %v", err)
	}
	client := NewClient(t.Context(), WithBaseURL(u), WithTokenProvider(staticTokenProvider("token-value")))

	result, err := client.ScoreData(t.Context(), "mod 1", "score step", ScoreDataRequest{Inputs: []ScoreInput{{Name: "age", Value: 35}}})
	if err != nil {
		t.Fatalf("ScoreData() error = %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestScoreDataRejectsEmptyIDs(t *testing.T) {
	client := NewClient(t.Context(), WithBaseURL(&url.URL{}), WithTokenProvider(staticTokenProvider("token")))

	_, err := client.ScoreData(t.Context(), "", "score", ScoreDataRequest{})
	if err == nil {
		t.Fatal("expected error for empty moduleID")
	}
	_, err = client.ScoreData(t.Context(), "mod", "", ScoreDataRequest{})
	if err == nil {
		t.Fatal("expected error for empty stepID")
	}
}
