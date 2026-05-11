package main

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
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

func TestResolveWorkflowContextPrefersCLIOverrides(t *testing.T) {
	doc := workflowDocument{
		Config: workflowProjectConfig{
			Defaults: workflowProjectDefaults{
				ContextID:   "workflow-id",
				ContextName: "workflow-name",
			},
		},
		User: workflowUserConfig{
			ContextID:   "user-id",
			ContextName: "user-name",
		},
	}
	cfg := cliConfig{
		ContextID:   "cli-id",
		ContextName: "cli-name",
	}

	contextID, contextName, err := resolveWorkflowContext(t.Context(), nil, cfg, doc, workflowFlagOverrides{contextID: true, contextName: true})
	if err != nil {
		t.Fatalf("resolveWorkflowContext() error = %v", err)
	}
	if got, want := contextID, "cli-id"; got != want {
		t.Fatalf("contextID = %q, want %q", got, want)
	}
	if got, want := contextName, "cli-name"; got != want {
		t.Fatalf("contextName = %q, want %q", got, want)
	}
}

func TestResolveWorkflowContextDoesNotUseConfiguredDefaultsAsCLIOverrides(t *testing.T) {
	doc := workflowDocument{
		Config: workflowProjectConfig{
			Defaults: workflowProjectDefaults{
				ContextID:   "workflow-id",
				ContextName: "workflow-name",
			},
		},
		User: workflowUserConfig{
			ContextID:   "user-id",
			ContextName: "user-name",
		},
	}
	// cfg values simulate ambient config loaded from env/profile, not explicit CLI flags.
	cfg := cliConfig{ContextID: "env-id", ContextName: "env-name"}

	contextID, contextName, err := resolveWorkflowContext(t.Context(), nil, cfg, doc, workflowFlagOverrides{})
	if err != nil {
		t.Fatalf("resolveWorkflowContext() error = %v", err)
	}
	if got, want := contextID, "workflow-id"; got != want {
		t.Fatalf("contextID = %q, want %q", got, want)
	}
	if got, want := contextName, "workflow-name"; got != want {
		t.Fatalf("contextName = %q, want %q", got, want)
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

func TestDataUploadCSVCommandReadsStdinAndWritesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers/server%201/caslibs/Public%20Data/tables"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm() error = %v", err)
		}
		if got, want := r.FormValue("tableName"), "class table"; got != want {
			t.Fatalf("tableName = %q, want %q", got, want)
		}
		content := readCLIMultipartFile(t, r.MultipartForm.File["file"][0])
		if got, want := string(content), "Name,Age\nAlice,13\n"; got != want {
			t.Fatalf("file content = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"class table","caslibName":"Public Data","rowCount":1,"columnCount":2}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLIWithStdin("Name,Age\nAlice,13\n",
		"data", "--base-url", server.URL, "--access-token", "test-token", "-o", "json",
		"upload-csv", "--server", "server 1", "--caslib", "Public Data", "--table", "class table", "--file", "-",
	)
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, `"ok": true`) || !strings.Contains(stdout, `"name": "class table"`) {
		t.Fatalf("stdout = %s, want uploaded table JSON", stdout)
	}
}

func TestDataPromoteCommandWritesText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/casManagement/servers/server%201/caslibs/Public%20Data/tables/class%20table"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"name":"class table","caslibName":"Public Data","scope":"global"}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLI("data", "--base-url", server.URL, "--access-token", "test-token",
		"promote", "--server", "server 1", "--caslib", "Public Data", "--table", "class table")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, "class table") || !strings.Contains(stdout, "global") {
		t.Fatalf("stdout = %s, want promoted table text", stdout)
	}
}

func TestFilesListCommandWritesJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/files/files?filter=contains%28name%2C%27report%27%29&limit=7"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"file-1","name":"report.txt","contentType":"text/plain","size":12}]}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLI("files", "--base-url", server.URL, "--access-token", "test-token", "-o", "json",
		"list", "--limit", "7", "--filter-name", "report")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, `"name": "report.txt"`) {
		t.Fatalf("stdout = %s, want file JSON", stdout)
	}
}

func TestFilesUploadCommandReadsStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Content-Disposition"), `attachment; filename="report.txt"`; got != want {
			t.Fatalf("Content-Disposition = %q, want %q", got, want)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		if got, want := string(body), "hello"; got != want {
			t.Fatalf("body = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"file-1","name":"report.txt","contentType":"text/plain","size":5}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLIWithStdin("hello", "files", "--base-url", server.URL, "--access-token", "test-token",
		"upload", "--name", "report.txt", "--file", "-", "--content-type", "text/plain")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, "report.txt") {
		t.Fatalf("stdout = %s, want file text", stdout)
	}
}

func TestFilesDownloadCommandWritesRawText(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/files/files/file%201/content"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		_, _ = w.Write([]byte("downloaded"))
	}))
	defer server.Close()

	stdout, _, err := executeCLI("files", "--base-url", server.URL, "--access-token", "test-token",
		"download", "--id", "file 1")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if got, want := stdout, "downloaded"; got != want {
		t.Fatalf("stdout = %q, want %q", got, want)
	}
}

func TestJobsSubmitCommandUsesConfiguredContextName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		var body struct {
			Name          string `json:"name"`
			JobDefinition struct {
				Code string `json:"code"`
			} `json:"jobDefinition"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got, want := body.JobDefinition.Code, "proc options; run;"; got != want {
			t.Fatalf("code = %q, want %q", got, want)
		}
		if got, want := body.Arguments["_contextName"], "ctx name"; got != want {
			t.Fatalf("_contextName = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"job-1","name":"job one","state":"submitted"}`))
	}))
	defer server.Close()

	stdout, _, err := executeCLI("jobs", "--base-url", server.URL, "--access-token", "test-token", "-o", "json",
		"submit", "--code", "proc options; run;", "--name", "job one", "--context-name", "ctx name")
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, `"id": "job-1"`) {
		t.Fatalf("stdout = %s, want job JSON", stdout)
	}
}

func TestJobsListStatusCancelAndLogCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.RequestURI == "/jobExecution/jobs?limit=2":
			_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"job-1","name":"job one","state":"completed"}]}`))
		case r.Method == http.MethodGet && r.RequestURI == "/jobExecution/jobs/job%201":
			_, _ = w.Write([]byte(`{"id":"job 1","name":"job one","state":"completed","results":{"main.log.txt":"/files/files/log-1"}}`))
		case r.Method == http.MethodDelete && r.RequestURI == "/jobExecution/jobs/job%201":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.RequestURI == "/files/files/log-1/content":
			_, _ = w.Write([]byte("job log"))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.RequestURI)
		}
	}))
	defer server.Close()

	stdout, _, err := executeCLI("jobs", "--base-url", server.URL, "--access-token", "test-token", "list", "--limit", "2")
	if err != nil || !strings.Contains(stdout, "job one") {
		t.Fatalf("list stdout=%s err=%v, want job one", stdout, err)
	}

	stdout, _, err = executeCLI("jobs", "--base-url", server.URL, "--access-token", "test-token", "status", "--id", "job 1")
	if err != nil || !strings.Contains(stdout, "completed") {
		t.Fatalf("status stdout=%s err=%v, want completed", stdout, err)
	}

	stdout, _, err = executeCLI("jobs", "--base-url", server.URL, "--access-token", "test-token", "cancel", "--id", "job 1")
	if err != nil || !strings.Contains(stdout, "cancelled") {
		t.Fatalf("cancel stdout=%s err=%v, want cancelled", stdout, err)
	}

	stdout, _, err = executeCLI("jobs", "--base-url", server.URL, "--access-token", "test-token", "log", "--id", "job 1")
	if err != nil || stdout != "job log\n" {
		t.Fatalf("log stdout=%q err=%v, want job log", stdout, err)
	}
}

func TestFilesCommandMissingFlagWritesFailureJSON(t *testing.T) {
	stdout, _, err := executeCLI("files", "--base-url", "https://viya.example.com", "--access-token", "test-token", "-o", "json", "download")
	if err == nil {
		t.Fatal("executeCLI() error = nil, want exit error")
	}
	if !strings.Contains(stdout, `"ok": false`) || !strings.Contains(stdout, "--id is required") {
		t.Fatalf("stdout = %s, want missing id failure", stdout)
	}
}

func TestWorkflowValidateAcceptsNestedArrayPlan(t *testing.T) {
	dir := t.TempDir()
	workflowPath := filepath.Join(dir, "workflow.yaml")
	if err := os.WriteFile(workflowPath, []byte(`version: 1
name: demo
steps:
  - name: prepare
    file: prepare.sas
  - - name: branch-a
      file: branch-a.sas
    - name: branch-b
      file: branch-b.sas
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	stdout, _, err := executeCLI("workflow", "-o", "json", "validate", "--file", workflowPath)
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, `"ok": true`) || !strings.Contains(stdout, "demo") {
		t.Fatalf("stdout = %s, want valid workflow JSON", stdout)
	}
}

func TestWorkflowRunUsesSingleComputeSessionForMultipleJobs(t *testing.T) {
	dir := t.TempDir()
	for name, content := range map[string]string{
		"prepare.sas":  "data _null_; put 'prepare'; run;",
		"branch-a.sas": "data _null_; put 'a'; run;",
		"branch-b.sas": "data _null_; put 'b'; run;",
	} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	workflowPath := filepath.Join(dir, "workflow.json")
	if err := os.WriteFile(workflowPath, []byte(`{
  "version": 1,
  "name": "demo-workflow",
  "maxParallel": 2,
  "contextId": "ctx-1",
  "includeOutput": true,
  "steps": [
    {"name": "prepare", "file": "prepare.sas", "log": "logs/prepare.log"},
    [
      {"name": "branch-a", "file": "branch-a.sas"},
      {"name": "branch-b", "file": "branch-b.sas"}
    ]
  ]
}`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	userConfigPath := filepath.Join(dir, "user.yaml")
	if err := os.WriteFile(userConfigPath, []byte(`autoexec: '%put autoexec;'
preCode: '%put pre;'
postCode: '%put post;'
variables:
  USER_LEVEL: yes
`), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	var mu sync.Mutex
	createdSessions := 0
	createdJobs := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/compute/contexts/ctx-1":
			_, _ = w.Write([]byte(`{"id":"ctx-1","name":"ctx one"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/compute/contexts/ctx-1/sessions":
			createdSessions++
			var body struct {
				Name        string `json:"name"`
				Environment struct {
					InitCode []string `json:"initCode"`
				} `json:"environment"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode session request: %v", err)
			}
			if got, want := body.Name, "demo-workflow"; got != want {
				t.Fatalf("session name = %q, want %q", got, want)
			}
			if len(body.Environment.InitCode) == 0 || !strings.Contains(body.Environment.InitCode[0], "autoexec") {
				t.Fatalf("initCode = %#v, want autoexec", body.Environment.InitCode)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"session-1","name":"demo-workflow","state":"idle"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/compute/sessions/session-1/jobs":
			createdJobs++
			var body struct {
				Name      string         `json:"name"`
				Code      []string       `json:"code"`
				Variables map[string]any `json:"variables"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode job request: %v", err)
			}
			joined := strings.Join(body.Code, "\n")
			if !strings.Contains(joined, "%put pre;") || !strings.Contains(joined, "%put post;") {
				t.Fatalf("job code = %q, want pre/post code", joined)
			}
			if body.Variables["WORKFLOW_STEP_PATH"] == "" || body.Variables["USER_LEVEL"] != "yes" {
				t.Fatalf("variables = %#v, want workflow paths and user variables", body.Variables)
			}
			if body.Variables["_SASPROGRAMFILE"] == "" || body.Variables["_SASPROGRAMDIR"] == "" {
				t.Fatalf("variables = %#v, want SAS program path variables", body.Variables)
			}
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"job-` + body.Name + `","sessionId":"session-1","state":"running"}`))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/compute/sessions/session-1/jobs/job-") && strings.HasSuffix(r.URL.Path, "/state"):
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("completed"))
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/compute/sessions/session-1/jobs/job-") && !strings.HasSuffix(r.URL.Path, "/state") && !strings.HasSuffix(r.URL.Path, "/log") && !strings.HasSuffix(r.URL.Path, "/listing"):
			_, _ = w.Write([]byte(`{"id":"job-info","sessionId":"session-1","state":"completed","jobConditionCode":0}`))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/log"):
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("workflow log"))
		case r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/listing"):
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("workflow listing"))
		case r.Method == http.MethodDelete && r.URL.Path == "/compute/sessions/session-1":
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.RequestURI)
		}
	}))
	defer server.Close()

	stdout, _, err := executeCLI("workflow", "--base-url", server.URL, "--access-token", "test-token", "--user-config", userConfigPath, "-o", "json", "run", "--file", workflowPath)
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if createdSessions != 1 {
		t.Fatalf("createdSessions = %d, want 1", createdSessions)
	}
	if createdJobs != 3 {
		t.Fatalf("createdJobs = %d, want 3", createdJobs)
	}
	if !strings.Contains(stdout, `"ok": true`) || !strings.Contains(stdout, `"kind": "parallel"`) {
		t.Fatalf("stdout = %s, want workflow result JSON", stdout)
	}
	if _, err := os.Stat(filepath.Join(dir, "logs", "prepare.log")); err != nil {
		t.Fatalf("expected log artifact: %v", err)
	}
}

func TestWorkflowRunUsesContextFromUserConfig(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "program.sas"), []byte("data _null_; put 'hello'; run;"), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}
	workflowPath := filepath.Join(dir, "workflow.yaml")
	if err := os.WriteFile(workflowPath, []byte(`version: 1
name: user-context
steps:
  - name: program
    file: program.sas
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}
	userConfigPath := filepath.Join(dir, "user.yaml")
	if err := os.WriteFile(userConfigPath, []byte(`contextName: user context
preCode: '%put user context;'
`), 0o644); err != nil {
		t.Fatalf("write user config: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/compute/contexts":
			_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"ctx-user","name":"user context"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/compute/contexts/ctx-user/sessions":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"session-user","name":"user-context","state":"idle"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/compute/sessions/session-user/jobs":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"job-user","sessionId":"session-user","state":"running"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/compute/sessions/session-user/jobs/job-user/state":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("completed"))
		case r.Method == http.MethodGet && r.URL.Path == "/compute/sessions/session-user/jobs/job-user":
			_, _ = w.Write([]byte(`{"id":"job-user","sessionId":"session-user","state":"completed","jobConditionCode":0}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/compute/sessions/session-user":
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.RequestURI)
		}
	}))
	defer server.Close()

	stdout, _, err := executeCLI("workflow", "--base-url", server.URL, "--access-token", "test-token", "--user-config", userConfigPath, "-o", "json", "run", "--file", workflowPath)
	if err != nil {
		t.Fatalf("executeCLI() error = %v, stdout = %s", err, stdout)
	}
	if !strings.Contains(stdout, `"contextId": "ctx-user"`) || !strings.Contains(stdout, `"contextName": "user context"`) {
		t.Fatalf("stdout = %s, want context from user config", stdout)
	}
}

func TestWorkflowRunFailsWhenComputeJobStateIsNotCompleted(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "program.sas"), []byte("data _null_; put 'fail'; run;"), 0o644); err != nil {
		t.Fatalf("write program: %v", err)
	}
	workflowPath := filepath.Join(dir, "workflow.yaml")
	if err := os.WriteFile(workflowPath, []byte(`version: 1
name: failed-state
steps:
  - name: only-step
    file: program.sas
`), 0o644); err != nil {
		t.Fatalf("write workflow: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/compute/contexts":
			_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"ctx-1","name":"SAS Job Execution compute context"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/compute/contexts/ctx-1/sessions":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"session-1","name":"failed-state","state":"idle"}`))
		case r.Method == http.MethodPost && r.URL.Path == "/compute/sessions/session-1/jobs":
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte(`{"id":"job-1","sessionId":"session-1","state":"running"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/compute/sessions/session-1/jobs/job-1/state":
			w.Header().Set("Content-Type", "text/plain")
			_, _ = w.Write([]byte("failed"))
		case r.Method == http.MethodGet && r.URL.Path == "/compute/sessions/session-1/jobs/job-1":
			_, _ = w.Write([]byte(`{"id":"job-1","sessionId":"session-1","state":"failed","jobConditionCode":5}`))
		case r.Method == http.MethodDelete && r.URL.Path == "/compute/sessions/session-1":
			w.WriteHeader(http.StatusAccepted)
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.RequestURI)
		}
	}))
	defer server.Close()

	stdout, _, err := executeCLI("workflow", "--base-url", server.URL, "--access-token", "test-token", "-o", "json", "run", "--file", workflowPath)
	if err == nil {
		t.Fatalf("executeCLI() error = nil, stdout = %s", stdout)
	}
	if !strings.Contains(stdout, `"ok": false`) || !strings.Contains(stdout, `compute job finished with state \"failed\"`) {
		t.Fatalf("stdout = %s, want failed state error", stdout)
	}
}

func executeCLI(args ...string) (stdout string, stderr string, err error) {
	return executeCLIWithStdin("", args...)
}

func executeCLIWithStdin(stdin string, args ...string) (stdout string, stderr string, err error) {
	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := newRootCommand(cliIO{
		stdout: &out,
		stderr: &errOut,
		stdin:  strings.NewReader(stdin),
	})
	cmd.SetArgs(args)
	err = cmd.Execute()
	return out.String(), errOut.String(), err
}

func readCLIMultipartFile(t *testing.T, header *multipart.FileHeader) []byte {
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
