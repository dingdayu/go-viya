package viya

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSendBatchJobInputSendsLines(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.RequestURI, "/batch/jobs/job%201/input"; got != want {
			t.Fatalf("request URI = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}
		if got := r.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/json") {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var req BatchJobInputRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !reflect.DeepEqual(req.Input, []string{"first\n", "second\n"}) {
			t.Fatalf("input = %#v, want %#v", req.Input, []string{"first\n", "second\n"})
		}
		if got, want := req.Version, 1; got != want {
			t.Fatalf("version = %d, want %d", got, want)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.SendBatchJobInput(context.Background(), "job 1", []string{"first\n", "second\n"})
	if err != nil {
		t.Fatalf("SendBatchJobInput() error = %v", err)
	}
}

func TestGetBatchJobOutputDecodesOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPost; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/batch/jobs/job-1/output"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/json, application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":["warn\n"],"output":["line one\n","line two\n"],"version":1}`))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	output, err := client.GetBatchJobOutput(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("GetBatchJobOutput() error = %v", err)
	}
	if !reflect.DeepEqual(output.Errors, []string{"warn\n"}) {
		t.Fatalf("Errors = %#v, want %#v", output.Errors, []string{"warn\n"})
	}
	if !reflect.DeepEqual(output.Output, []string{"line one\n", "line two\n"}) {
		t.Fatalf("Output = %#v, want %#v", output.Output, []string{"line one\n", "line two\n"})
	}
	if got, want := output.Version, 1; got != want {
		t.Fatalf("Version = %d, want %d", got, want)
	}
}

func TestGetBatchJobStateReturnsPlainTextState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodGet; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/batch/jobs/job-1/state"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "text/plain, application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("completed\n"))
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	state, err := client.GetBatchJobState(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("GetBatchJobState() error = %v", err)
	}
	if got, want := state, "completed"; got != want {
		t.Fatalf("state = %q, want %q", got, want)
	}
}

func TestCancelBatchJobSetsCanceledState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got, want := r.Method, http.MethodPut; got != want {
			t.Fatalf("method = %q, want %q", got, want)
		}
		if got, want := r.URL.Path, "/batch/jobs/job-1/state"; got != want {
			t.Fatalf("path = %q, want %q", got, want)
		}
		if got, want := r.URL.Query().Get("value"), "canceled"; got != want {
			t.Fatalf("value query = %q, want %q", got, want)
		}
		if got, want := r.Header.Get("Accept"), "application/vnd.sas.error+json"; got != want {
			t.Fatalf("Accept = %q, want %q", got, want)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	err := client.CancelBatchJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("CancelBatchJob() error = %v", err)
	}
}

func TestWaitBatchJobCompletedReturnsFinalJobDetails(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/batch/jobs/job-1" {
			t.Fatalf("path = %q, want /batch/jobs/job-1", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		switch calls.Add(1) {
		case 1:
			_, _ = w.Write([]byte(`{"id":"job-1","name":"daily-load","state":"running","returnCode":0}`))
		case 2:
			_, _ = w.Write([]byte(`{"id":"job-1","name":"daily-load","state":"completed","returnCode":0,"logFile":"daily-load.log"}`))
		default:
			t.Fatalf("unexpected request count = %d", calls.Load())
		}
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	jobInfo, err := client.WaitBatchJobCompleted(context.Background(), "job-1", time.Millisecond)
	if err != nil {
		t.Fatalf("WaitBatchJobCompleted() error = %v", err)
	}
	if got, want := jobInfo.ID, "job-1"; got != want {
		t.Fatalf("jobInfo.ID = %q, want %q", got, want)
	}
	if got, want := jobInfo.Name, "daily-load"; got != want {
		t.Fatalf("jobInfo.Name = %q, want %q", got, want)
	}
	if got, want := jobInfo.State, "completed"; got != want {
		t.Fatalf("jobInfo.State = %q, want %q", got, want)
	}
	if got, want := jobInfo.LogFile, "daily-load.log"; got != want {
		t.Fatalf("jobInfo.LogFile = %q, want %q", got, want)
	}
	if got, want := calls.Load(), int32(2); got != want {
		t.Fatalf("requests = %d, want %d", got, want)
	}
}

func TestWaitBatchJobCompletedReturnsMostRecentJobOnError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch calls.Add(1) {
		case 1:
			_, _ = w.Write([]byte(`{"id":"job-1","name":"daily-load","state":"running","returnCode":0}`))
		case 2:
			http.Error(w, `{"message":"unavailable"}`, http.StatusServiceUnavailable)
		default:
			t.Fatalf("unexpected request count = %d", calls.Load())
		}
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)

	jobInfo, err := client.WaitBatchJobCompleted(context.Background(), "job-1", time.Millisecond)
	if err == nil {
		t.Fatal("WaitBatchJobCompleted() error = nil, want error")
	}
	if !strings.Contains(err.Error(), "status code: 503") {
		t.Fatalf("error = %q, want status code 503", err.Error())
	}
	if got, want := jobInfo.State, "running"; got != want {
		t.Fatalf("jobInfo.State = %q, want most recent state %q", got, want)
	}
	if got, want := jobInfo.Name, "daily-load"; got != want {
		t.Fatalf("jobInfo.Name = %q, want %q", got, want)
	}
}

func TestWaitBatchJobCompletedReturnsContextError(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewClient(context.Background(), server.URL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	jobInfo, err := client.WaitBatchJobCompleted(ctx, "job-1", time.Hour)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("WaitBatchJobCompleted() error = %v, want context.Canceled", err)
	}
	if !reflect.DeepEqual(jobInfo, BatchJob{}) {
		t.Fatalf("jobInfo = %+v, want zero BatchJob", jobInfo)
	}
	if got := calls.Load(); got != 0 {
		t.Fatalf("requests = %d, want 0", got)
	}
}
