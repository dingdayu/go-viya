package viya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSubmitJobExecutionCodeSendsBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/jobExecution/jobs"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Authorization"), "Bearer token-value"; got != want {
			t.Fatalf("Authorization = %q, want %q", got, want)
		}

		var body struct {
			Name          string `json:"name"`
			JobDefinition struct {
				Type string `json:"type"`
				Code string `json:"code"`
			} `json:"jobDefinition"`
			Arguments map[string]string `json:"arguments"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if got, want := body.Name, "job one"; got != want {
			t.Fatalf("name = %q, want %q", got, want)
		}
		if got, want := body.JobDefinition.Type, "Compute"; got != want {
			t.Fatalf("type = %q, want %q", got, want)
		}
		if got, want := body.JobDefinition.Code, "proc options; run;"; got != want {
			t.Fatalf("code = %q, want %q", got, want)
		}
		if got, want := body.Arguments["_contextName"], "SAS Job Execution compute context"; got != want {
			t.Fatalf("_contextName = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"job-1","name":"job one","state":"submitted"}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL, WithTokenProvider(staticTokenProvider("token-value")))

	job, err := client.SubmitJobExecutionCode(context.Background(), SubmitJobExecutionCodeRequest{
		Name:        "job one",
		Code:        "proc options; run;",
		ContextName: "SAS Job Execution compute context",
	})
	if err != nil {
		t.Fatalf("SubmitJobExecutionCode() error = %v", err)
	}
	if got, want := job.ID, "job-1"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestGetJobExecutionJobsSetsPaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/jobExecution/jobs?limit=20&start=3"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"job-1","name":"job one","state":"completed"}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	jobs, err := client.GetJobExecutionJobs(context.Background(), ListOptions{Start: 3, Limit: 20})
	if err != nil {
		t.Fatalf("GetJobExecutionJobs() error = %v", err)
	}
	if got, want := jobs.Items[0].State, "completed"; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
}

func TestGetAndCancelJobExecutionJobEscapeID(t *testing.T) {
	var sawGet bool
	var sawDelete bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.RequestURI, "/jobExecution/jobs/job%201"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		switch r.Method {
		case http.MethodGet:
			sawGet = true
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"id":"job 1","state":"running"}`))
		case http.MethodDelete:
			sawDelete = true
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Fatalf("method = %q, want GET or DELETE", r.Method)
		}
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	job, err := client.GetJobExecutionJob(context.Background(), "job 1")
	if err != nil {
		t.Fatalf("GetJobExecutionJob() error = %v", err)
	}
	if got, want := job.State, "running"; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
	if err := client.CancelJobExecutionJob(context.Background(), "job 1"); err != nil {
		t.Fatalf("CancelJobExecutionJob() error = %v", err)
	}
	if !sawGet || !sawDelete {
		t.Fatalf("sawGet=%t sawDelete=%t, want both true", sawGet, sawDelete)
	}
}

func TestGetJobExecutionJobLogPrefersLogTxt(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.RequestURI {
		case "/jobExecution/jobs/job-1":
			_, _ = w.Write([]byte(`{"id":"job-1","state":"completed","results":{"a.log":"/files/files/log-a","b.log.txt":"/files/files/log-b"}}`))
		case "/files/files/log-b/content":
			_, _ = w.Write([]byte("preferred log"))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.RequestURI)
		}
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	log, err := client.GetJobExecutionJobLog(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("GetJobExecutionJobLog() error = %v", err)
	}
	if got, want := log, "preferred log"; got != want {
		t.Fatalf("log = %q, want %q", got, want)
	}
}

func TestGetJobExecutionJobLogFallsBackToLogAndNoLogMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.RequestURI {
		case "/jobExecution/jobs/job-log":
			_, _ = w.Write([]byte(`{"id":"job-log","state":"completed","results":{"main.log":"/files/files/main-log"}}`))
		case "/files/files/main-log/content":
			_, _ = w.Write([]byte("fallback log"))
		case "/jobExecution/jobs/job-no-log":
			_, _ = w.Write([]byte(`{"id":"job-no-log","state":"failed","error":{"message":"bad code"}}`))
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.RequestURI)
		}
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	log, err := client.GetJobExecutionJobLog(context.Background(), "job-log")
	if err != nil {
		t.Fatalf("GetJobExecutionJobLog(job-log) error = %v", err)
	}
	if got, want := log, "fallback log"; got != want {
		t.Fatalf("log = %q, want %q", got, want)
	}

	log, err = client.GetJobExecutionJobLog(context.Background(), "job-no-log")
	if err != nil {
		t.Fatalf("GetJobExecutionJobLog(job-no-log) error = %v", err)
	}
	if !strings.Contains(log, "Job failed: bad code") {
		t.Fatalf("log = %q, want failure details", log)
	}
}
