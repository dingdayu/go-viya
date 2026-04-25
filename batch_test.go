package viya

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

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
