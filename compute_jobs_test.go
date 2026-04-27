package viya

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestCreateComputeJobSendsCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/compute/sessions/session%201/jobs"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}

		var req CreateComputeJobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(req.Code, []string{"data _null_;", "run;"}) {
			t.Fatalf("Code = %#v, want %#v", req.Code, []string{"data _null_;", "run;"})
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"0","sessionId":"session 1","state":"running","version":1}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	job, err := client.CreateComputeJob(context.Background(), "session 1", CreateComputeJobRequest{
		Code: []string{"data _null_;", "run;"},
	})
	if err != nil {
		t.Fatalf("CreateComputeJob() error = %v", err)
	}
	if got, want := job.ID, "0"; got != want {
		t.Fatalf("ID = %q, want %q", got, want)
	}
}

func TestGetComputeJobsListDecodesJobs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.URL.Path, "/compute/sessions/session-1/jobs"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"count":1,"items":[{"id":"0","sessionId":"session-1","state":"completed","jobConditionCode":0}]}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	jobs, err := client.GetComputeJobsList(context.Background(), "session-1")
	if err != nil {
		t.Fatalf("GetComputeJobsList() error = %v", err)
	}
	if got, want := jobs.Count, 1; got != want {
		t.Fatalf("Count = %d, want %d", got, want)
	}
	if got, want := jobs.Items[0].State, "completed"; got != want {
		t.Fatalf("State = %q, want %q", got, want)
	}
}

func TestCancelComputeJobSendsIfMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/compute/sessions/session-1/jobs/0/state"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("value"), "canceled"; got != want {
			t.Fatalf("value query = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("If-Match"), `"job-etag"`; got != want {
			t.Fatalf("If-Match = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("canceled\n"))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	state, err := client.CancelComputeJob(context.Background(), "session-1", "0", `"job-etag"`)
	if err != nil {
		t.Fatalf("CancelComputeJob() error = %v", err)
	}
	if got, want := state, "canceled"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestGetComputeJobStateReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"missing job"}`, http.StatusNotFound)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	_, err := client.GetComputeJobState(context.Background(), "session-1", "0")
	if err == nil {
		t.Fatal("GetComputeJobState() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "status code: 404") {
		t.Fatalf("error = %q, want status code 404", err.Error())
	}
}
